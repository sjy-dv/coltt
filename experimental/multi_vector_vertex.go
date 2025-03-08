// Licensed to sjy-dv under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. sjy-dv licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package experimental

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"sync/atomic"

	"github.com/sjy-dv/coltt/gen/protoc/v3/experimentalproto"
	"github.com/sjy-dv/coltt/pkg/distance"
	"github.com/sjy-dv/coltt/pkg/sharding"
)

type multiVectorVertex struct {
	vertexMetadata Metadata
	collectionName string
	size           uint64
	vertices       [VERTEX_SHARD_COUNT]map[string]VertexEdge
	verticesMu     [VERTEX_SHARD_COUNT]*sync.RWMutex
	distance       distance.Space
}

func newMultiVectorVertex(collectionName string, metadata Metadata) *multiVectorVertex {
	vecspace := &multiVectorVertex{
		vertexMetadata: metadata,
		collectionName: collectionName,
		distance: func() distance.Space {
			if metadata.Distancer() == experimentalproto.Distance_Cosine {
				return distance.NewCosine()
			}
			return distance.NewEuclidean()
		}(),
	}
	for i := 0; i < VERTEX_SHARD_COUNT; i++ {
		vecspace.vertices[i] = make(map[string]VertexEdge)
		vecspace.verticesMu[i] = &sync.RWMutex{}
	}
	return vecspace
}

func (vertex *multiVectorVertex) ChangedVertex(Id string, edge VertexEdge) error {
	for key, vec := range edge.MultiVectors {
		if vertex.vertexMetadata.Dimensional() != uint32(vec.Dimensions()) {
			return fmt.Errorf("index [%s] expect dimension: [%d], but got [%d]", key, vertex.vertexMetadata.Dimensional(), vec.Dimensions())
		}
		if vertex.distance.Type() == T_COSINE {
			edge.MultiVectors[key] = Normalize(vec)
		}
	}
	shardArea := sharding.ShardVertexV2(Id, uint64(VERTEX_SHARD_COUNT))
	vertex.verticesMu[shardArea].Lock()
	defer vertex.verticesMu[shardArea].Unlock()
	vertex.vertices[shardArea][Id] = edge
	return nil
}

func (vertex *multiVectorVertex) RemoveVertex(Id string) error {
	shardArea := sharding.ShardVertexV2(Id, uint64(VERTEX_SHARD_COUNT))
	vertex.verticesMu[shardArea].Lock()
	defer vertex.verticesMu[shardArea].Unlock()
	delete(vertex.vertices[shardArea], Id)
	return nil
}

func (vertex *multiVectorVertex) MultiVertexSearch(topK uint64, multiVectors []*experimentalproto.MultiVectorIndex) (
	[]*NearestNeighbor, error,
) {
	for idx, vectors := range multiVectors {
		c, ok := vertex.vertexMetadata.IndexType[vectors.GetIndexName()]
		if !ok || experimentalproto.IndexType(c.IndexType) != experimentalproto.IndexType_Vector {
			return nil, fmt.Errorf("index [%s] is not defined vector fields", vectors.GetIndexName())
		}
		if vertex.vertexMetadata.Dimensional() != uint32(len(vectors.Vector)) {
			return nil, fmt.Errorf("index [%s] expect dimension: [%d], but got [%d]", vectors.GetIndexName(), vertex.vertexMetadata.Dimensional(), len(vectors.Vector))
		}
		if vectors.IncludeOrNot {
			if vertex.distance.Type() == T_COSINE {
				multiVectors[idx].Vector = Normalize(vectors.GetVector())
			}
		}
	}
	pq := NewPriorityQueue(int(topK))
	results := make([]shardNeighbor, VERTEX_SHARD_COUNT)
	var shardManager sync.WaitGroup
	shardManager.Add(VERTEX_SHARD_COUNT)
	for shard := 0; shard < VERTEX_SHARD_COUNT; shard++ {
		go func(shard int) {
			defer shardManager.Done()
			localpq := NewPriorityQueue(int(topK))
			vertex.verticesMu[shard].RLock()
			for pk, node := range vertex.vertices[shard] {
				score := float32(0)
				for _, vectors := range multiVectors {
					if vectors.IncludeOrNot {
						sim := vertex.distance.Distance(node.MultiVectors[vectors.IndexName], vectors.GetVector())
						score += (scoreHelper(sim, vertex.distance.Type()) * (float32(vectors.Ratio) / 100))
					}
				}
				localpq.Add(&NearestNeighbor{
					Id:       pk,
					Metadata: node.Metadata,
					Score:    score,
				})
			}
			vertex.verticesMu[shard].RUnlock()
			results[shard].NN = localpq.ToSlice()
		}(shard)
	}
	shardManager.Wait()
	for _, res := range results {
		for _, item := range res.NN {
			pq.Add(item)
		}
	}

	return pq.ToSlice(), nil
}

func (vertex *multiVectorVertex) SaveVertexMetadata() ([]byte, error) {
	return json.Marshal(vertex.vertexMetadata)
}

func (vertex *multiVectorVertex) LoadVertexMetadata(collectionName string, data []byte) error {
	var metadata Metadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return err
	}
	vertex.collectionName = collectionName
	vertex.vertexMetadata = metadata
	vertex.distance = func() distance.Space {
		if metadata.Distancer() == experimentalproto.Distance_Cosine {
			return distance.NewCosine()
		}
		return distance.NewEuclidean()
	}()
	return nil
}

func (vertex *multiVectorVertex) Quantization() experimentalproto.Quantization {
	return vertex.vertexMetadata.Quantizationer()
}

func (vertex *multiVectorVertex) Distance() experimentalproto.Distance {
	return vertex.vertexMetadata.Distancer()
}

func (vertex *multiVectorVertex) Dim() uint32 {
	return vertex.vertexMetadata.Dimensional()
}

func (vertex *multiVectorVertex) LoadSize() int64 {
	return int64(atomic.LoadUint64(&vertex.size))
}

func (vertex *multiVectorVertex) Indexer() map[string]IndexFeature {
	return vertex.vertexMetadata.IndexType
}

func (vertex *multiVectorVertex) Versional() bool {
	return vertex.vertexMetadata.Versional()
}

func (vertex *multiVectorVertex) SaveVertex() ([]byte, error) {
	var buf bytes.Buffer

	for i := 0; i < VERTEX_SHARD_COUNT; i++ {
		vertex.verticesMu[i].RLock()
		entries := vertex.vertices[i]
		if err := binary.Write(&buf, binary.BigEndian, uint64(len(entries))); err != nil {
			return nil, err
		}
		for key, value := range entries {
			keyBytes := []byte(key)
			if len(keyBytes) > 65535 {
				return nil, fmt.Errorf("key too long: %s", key)
			}
			if err := binary.Write(&buf, binary.BigEndian, uint16(len(keyBytes))); err != nil {
				return nil, err
			}
			if _, err := buf.Write(keyBytes); err != nil {
				return nil, err
			}
			if err := binary.Write(&buf, binary.BigEndian, uint32(len(value.MultiVectors))); err != nil {
				return nil, err
			}
			for mvKey, vec := range value.MultiVectors {
				// multiVector key
				mvKeyBytes := []byte(mvKey)
				if len(mvKeyBytes) > 65535 {
					return nil, fmt.Errorf("multiVector key too long: %s", mvKey)
				}
				if err := binary.Write(&buf, binary.BigEndian, uint16(len(mvKeyBytes))); err != nil {
					return nil, err
				}
				if _, err := buf.Write(mvKeyBytes); err != nil {
					return nil, err
				}

				if err := binary.Write(&buf, binary.BigEndian, uint32(len(vec))); err != nil {
					return nil, err
				}
				for _, f := range vec {
					if err := binary.Write(&buf, binary.BigEndian, f); err != nil {
						return nil, err
					}
				}
			}
			if err := binary.Write(&buf, binary.BigEndian, uint32(len(value.Metadata))); err != nil {
				return nil, err
			}
			for metaKey, metaVal := range value.Metadata {
				// metadata key: len(uint16) + byte
				metaKeyBytes := []byte(metaKey)
				if len(metaKeyBytes) > 65535 {
					return nil, fmt.Errorf("metadata key too long: %s", metaKey)
				}
				if err := binary.Write(&buf, binary.BigEndian, uint16(len(metaKeyBytes))); err != nil {
					return nil, err
				}
				if _, err := buf.Write(metaKeyBytes); err != nil {
					return nil, err
				}

				// metadata : type tag(1byte) + value
				switch v := metaVal.(type) {
				case int64:
					// tag 0: int64
					if err := buf.WriteByte(0); err != nil {
						return nil, err
					}
					if err := binary.Write(&buf, binary.BigEndian, v); err != nil {
						return nil, err
					}
				case string:
					//tag 1: string
					if err := buf.WriteByte(1); err != nil {
						return nil, err
					}
					strBytes := []byte(v)
					if len(strBytes) > 65535 {
						return nil, fmt.Errorf("metadata string too long: %s", v)
					}
					if err := binary.Write(&buf, binary.BigEndian, uint16(len(strBytes))); err != nil {
						return nil, err
					}
					if _, err := buf.Write(strBytes); err != nil {
						return nil, err
					}
				default:
					return nil, fmt.Errorf("unsupported metadata type: %T", v)
				}
			}
		}
		vertex.verticesMu[i].RUnlock()
	}
	return buf.Bytes(), nil
}

func (vertex *multiVectorVertex) LoadVertex(data []byte) error {
	var vertices [VERTEX_SHARD_COUNT]map[string]VertexEdge
	buf := bytes.NewReader(data)

	for i := 0; i < VERTEX_SHARD_COUNT; i++ {
		var count uint64
		if err := binary.Read(buf, binary.BigEndian, &count); err != nil {
			return err
		}
		atomic.AddUint64(&vertex.size, count)
		m := make(map[string]VertexEdge, count)
		for j := uint64(0); j < count; j++ {
			var keyLen uint16
			if err := binary.Read(buf, binary.BigEndian, &keyLen); err != nil {
				return err
			}
			keyBytes := make([]byte, keyLen)
			if _, err := io.ReadFull(buf, keyBytes); err != nil {
				return err
			}
			key := string(keyBytes)

			var ve VertexEdge

			var mvCount uint32
			if err := binary.Read(buf, binary.BigEndian, &mvCount); err != nil {
				return err
			}
			ve.MultiVectors = make(map[string]Vector, mvCount)
			for k := uint32(0); k < mvCount; k++ {
				var mvKeyLen uint16
				if err := binary.Read(buf, binary.BigEndian, &mvKeyLen); err != nil {
					return err
				}
				mvKeyBytes := make([]byte, mvKeyLen)
				if _, err := io.ReadFull(buf, mvKeyBytes); err != nil {
					return err
				}
				mvKey := string(mvKeyBytes)

				var dim uint32
				if err := binary.Read(buf, binary.BigEndian, &dim); err != nil {
					return err
				}
				vec := make(Vector, dim)
				for d := 0; d < int(dim); d++ {
					if err := binary.Read(buf, binary.BigEndian, &vec[d]); err != nil {
						return err
					}
				}
				ve.MultiVectors[mvKey] = vec
			}

			var metaCount uint32
			if err := binary.Read(buf, binary.BigEndian, &metaCount); err != nil {
				return err
			}
			ve.Metadata = make(map[string]any, metaCount)
			for k := uint32(0); k < metaCount; k++ {
				var metaKeyLen uint16
				if err := binary.Read(buf, binary.BigEndian, &metaKeyLen); err != nil {
					return err
				}
				metaKeyBytes := make([]byte, metaKeyLen)
				if _, err := io.ReadFull(buf, metaKeyBytes); err != nil {
					return err
				}
				metaKey := string(metaKeyBytes)

				typ, err := buf.ReadByte()
				if err != nil {
					return err
				}
				switch typ {
				case 0: // int64
					var val int64
					if err := binary.Read(buf, binary.BigEndian, &val); err != nil {
						return err
					}
					ve.Metadata[metaKey] = val
				case 1: // string
					var strLen uint16
					if err := binary.Read(buf, binary.BigEndian, &strLen); err != nil {
						return err
					}
					strBytes := make([]byte, strLen)
					if _, err := io.ReadFull(buf, strBytes); err != nil {
						return err
					}
					ve.Metadata[metaKey] = string(strBytes)
				default:
					return fmt.Errorf("unsupported metadata type tag: %d", typ)
				}
			}
			m[key] = ve
		}
		vertices[i] = m
	}
	for i := 0; i < VERTEX_SHARD_COUNT; i++ {
		vertex.vertices[i] = vertices[i]
		vertex.verticesMu[i] = &sync.RWMutex{}
	}
	return nil
}
