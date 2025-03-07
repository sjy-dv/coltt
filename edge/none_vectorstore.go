package edge

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"sync/atomic"

	"github.com/sjy-dv/coltt/gen/protoc/v4/edgepb"
	"github.com/sjy-dv/coltt/pkg/distance"
	"github.com/sjy-dv/coltt/pkg/inverted"
	"github.com/sjy-dv/coltt/pkg/sharding"
)

type noneVecSpace struct {
	vertexMetadata Metadata
	collectionName string
	size           uint64
	vertices       [EDGE_MAP_SHARD_COUNT]map[uint64]ENode
	verticesMu     [EDGE_MAP_SHARD_COUNT]*sync.RWMutex
	distance       distance.Space
	quantization   NoQuantization
	invertedIndex  *inverted.BitmapIndex
}

func newNoneVectorstore(collectionName string, metadata Metadata) *noneVecSpace {
	vecspace := &noneVecSpace{
		vertexMetadata: metadata,
		collectionName: collectionName,
		distance: func() distance.Space {
			if metadata.Distancer() == edgepb.Distance_Cosine {
				return distance.NewCosine()
			}
			return distance.NewEuclidean()
		}(),
		quantization:  NoQuantization{},
		invertedIndex: inverted.NewBitmapIndex(),
	}
	for i := 0; i < EDGE_MAP_SHARD_COUNT; i++ {
		vecspace.vertices[i] = make(map[uint64]ENode)
		vecspace.verticesMu[i] = &sync.RWMutex{}
	}
	return vecspace
}

func (vertex *noneVecSpace) ChangedVertex(updateId string, commitId uint64, data ENode) error {
	if updateId != "" {
		var primaryIndex string
		for _, indexer := range vertex.Indexer() {
			if indexer.PrimaryKey {
				primaryIndex = indexer.IndexName
				break
			}
		}
		finder := inverted.NewFilter(primaryIndex, inverted.OpEqual, updateId)
		ids, err := vertex.invertedIndex.SearchSingleFilter(finder)
		if err != nil {
			return err
		}
		if len(ids) != 0 {
			//없다면 생성
			//있다면 기존 id로 덮어씌우기
			commitId = ids[0]
		}
	}
	if vertex.vertexMetadata.Dimensional() != uint32(data.Vector.Dimensions()) {
		return fmt.Errorf("Dim Length UnmatchdError: expect dimension: [%d], but got [%d]", vertex.vertexMetadata.Dimensional(), data.Vector.Dimensions())
	}
	if err := standardAnalyzer(data.Metadata, vertex.Indexer()); err != nil {
		return err
	}
	if err := vertex.invertedIndex.Add(commitId, data.Metadata); err != nil {
		return fmt.Errorf("ErrInvertedIndexAddFailed: %s", err.Error())
	}
	if vertex.distance.Type() == T_COSINE {
		data.Vector = Normalize(data.Vector)
	}
	shardIdx := sharding.ShardVertex(commitId, uint64(EDGE_MAP_SHARD_COUNT))
	vertex.verticesMu[shardIdx].Lock()
	defer vertex.verticesMu[shardIdx].Unlock()
	vertex.vertices[shardIdx][commitId] = data
	return nil
}

func (vertex *noneVecSpace) RemoveVertex(dropFilter map[string]interface{}) error {
	if err := dropKeyAnalyzer(dropFilter, vertex.Indexer()); err != nil {
		return err
	}
	//해당 조건 모두 찾음
	filters := make([]*inverted.Filter, 0)
	for index, indexValue := range dropFilter {
		filter := inverted.NewFilter(index, inverted.OpEqual, indexValue)
		filters = append(filters, filter)
	}
	dropIds, err := vertex.invertedIndex.SearchMultiFilter(filters)
	if err != nil {
		return fmt.Errorf("InvertedIndexFindDeleteIdsError: %s", err.Error())
	}
	for _, id := range dropIds {
		shardIdx := sharding.ShardVertex(id, uint64(EDGE_MAP_SHARD_COUNT))
		vertex.verticesMu[shardIdx].Lock()
		vertex.invertedIndex.Remove(id, vertex.vertices[shardIdx][id].Metadata)
		delete(vertex.vertices[shardIdx], id)
		vertex.verticesMu[shardIdx].Unlock()
	}
	return nil
}

func (vertex *noneVecSpace) VertexSearch(target Vector, topK int, highCpu bool,
) ([]*SearchResultItem, error) {
	if vertex.distance.Type() == T_COSINE {
		target = Normalize(target)
	}
	pq := NewPriorityQueue(topK)
	if !highCpu {
		for shard := 0; shard < EDGE_MAP_SHARD_COUNT; shard++ {
			vertex.verticesMu[shard].RLock()
			for uid, node := range vertex.vertices[shard] {
				sim := vertex.quantization.Similarity(target, node.Vector, vertex.distance)
				pq.Add(&SearchResultItem{
					Id:       uid,
					Score:    sim,
					Metadata: node.Metadata,
				})
			}
			vertex.verticesMu[shard].RUnlock()
		}
	} else {
		type shardResult struct {
			Items []*SearchResultItem
		}
		results := make([]shardResult, EDGE_MAP_SHARD_COUNT)
		var concurrenyWorker sync.WaitGroup
		concurrenyWorker.Add(EDGE_MAP_SHARD_COUNT)
		for shard := 0; shard < EDGE_MAP_SHARD_COUNT; shard++ {
			go func(shard int) {
				defer concurrenyWorker.Done()
				localpq := NewPriorityQueue(topK)
				vertex.verticesMu[shard].RLock()
				for uid, node := range vertex.vertices[shard] {
					sim := vertex.quantization.Similarity(target, node.Vector, vertex.distance)
					localpq.Add(&SearchResultItem{
						Id:       uid,
						Score:    sim,
						Metadata: node.Metadata,
					})
				}
				vertex.verticesMu[shard].RUnlock()
				results[shard].Items = localpq.ToSlice()
			}(shard)
		}
		concurrenyWorker.Wait()
		for _, res := range results {
			for _, item := range res.Items {
				pq.Add(item)
			}
		}
	}
	return pq.ToSlice(), nil
}

func (vertex *noneVecSpace) FilterableVertexSearch(filter *inverted.FilterExpression, target Vector, topK int, highCpu bool) (
	[]*SearchResultItem, error,
) {
	if vertex.distance.Type() == T_COSINE {
		target = Normalize(target)
	}

	candidates, err := vertex.invertedIndex.SearchWithExpression(filter)
	if err != nil {
		return nil, err
	}
	shardCandidates := make([][]uint64, EDGE_MAP_SHARD_COUNT)
	for _, cand := range candidates {
		shardIndex := sharding.ShardVertex(cand, uint64(EDGE_MAP_SHARD_COUNT))
		shardCandidates[shardIndex] = append(shardCandidates[shardIndex], cand)
	}
	globalPQ := NewPriorityQueue(topK)
	if !highCpu {
		for shard := 0; shard < EDGE_MAP_SHARD_COUNT; shard++ {
			if len(shardCandidates[shard]) == 0 {
				continue
			}
			vertex.verticesMu[shard].RLock()
			for _, uid := range shardCandidates[shard] {
				if node, ok := vertex.vertices[shard][uid]; ok {
					sim := vertex.quantization.Similarity(target, node.Vector, vertex.distance)
					globalPQ.Add(&SearchResultItem{
						Id:       uid,
						Score:    sim,
						Metadata: node.Metadata,
					})
				}
			}
			vertex.verticesMu[shard].RUnlock()
		}
	} else {
		type shardResult struct {
			Items []*SearchResultItem
		}
		results := make([]shardResult, EDGE_MAP_SHARD_COUNT)
		var wg sync.WaitGroup
		wg.Add(EDGE_MAP_SHARD_COUNT)
		for shard := 0; shard < EDGE_MAP_SHARD_COUNT; shard++ {
			go func(shard int) {
				defer wg.Done()
				localPQ := NewPriorityQueue(topK)
				if len(shardCandidates[shard]) == 0 {
					return
				}
				vertex.verticesMu[shard].RLock()
				for _, uid := range shardCandidates[shard] {
					if node, ok := vertex.vertices[shard][uid]; ok {
						sim := vertex.quantization.Similarity(target, node.Vector, vertex.distance)
						localPQ.Add(&SearchResultItem{
							Id:       uid,
							Score:    sim,
							Metadata: node.Metadata,
						})
					}
				}
				vertex.verticesMu[shard].RUnlock()
				results[shard].Items = localPQ.ToSlice()
			}(shard)
		}
		wg.Wait()
		for _, res := range results {
			for _, item := range res.Items {
				globalPQ.Add(item)
			}
		}
	}
	return globalPQ.ToSlice(), nil
}

func (vertex *noneVecSpace) SaveVertexMetadata() ([]byte, error) {
	return json.Marshal(vertex.vertexMetadata)
}

func (vertex *noneVecSpace) LoadVertexMetadata(collectionName string, data []byte) error {
	var metadata Metadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return err
	}
	vertex.collectionName = collectionName
	vertex.vertexMetadata = metadata
	vertex.distance = func() distance.Space {
		if metadata.Distancer() == edgepb.Distance_Cosine {
			return distance.NewCosine()
		}
		return distance.NewEuclidean()
	}()
	return nil
}

func (vertex *noneVecSpace) SaveVertexInverted() ([]byte, error) {
	return vertex.invertedIndex.SerializeBinary()
}

func (vertex *noneVecSpace) LoadVertexInverted(data []byte) error {
	vertex.invertedIndex = inverted.NewBitmapIndex()
	return vertex.invertedIndex.DeserializeBinary(data)
}

func (vertex *noneVecSpace) Quantization() edgepb.Quantization {
	return vertex.vertexMetadata.Quantizationer()
}

func (vertex *noneVecSpace) Distance() edgepb.Distance {
	return vertex.vertexMetadata.Distancer()
}

func (vertex *noneVecSpace) Dim() uint32 {
	return vertex.vertexMetadata.Dimensional()
}

func (vertex *noneVecSpace) LoadSize() int64 {
	return int64(atomic.LoadUint64(&vertex.size))
}

func (vertex *noneVecSpace) Indexer() map[string]IndexFeature {
	return vertex.vertexMetadata.IndexType
}

func (vertex *noneVecSpace) Versional() bool {
	return vertex.vertexMetadata.Versional()
}

func (n *noneVecSpace) SaveVertex() ([]byte, error) {
	var buf bytes.Buffer

	for i := 0; i < EDGE_MAP_SHARD_COUNT; i++ {
		n.verticesMu[i].RLock()
		entries := n.vertices[i]
		if err := binary.Write(&buf, binary.BigEndian, uint64(len(entries))); err != nil {
			n.verticesMu[i].RUnlock()
			return nil, err
		}
		for key, node := range entries {
			if err := binary.Write(&buf, binary.BigEndian, key); err != nil {
				n.verticesMu[i].RUnlock()
				return nil, err
			}

			vecLen := uint32(len(node.Vector))
			if err := binary.Write(&buf, binary.BigEndian, vecLen); err != nil {
				n.verticesMu[i].RUnlock()
				return nil, err
			}
			for _, f := range node.Vector {
				if err := binary.Write(&buf, binary.BigEndian, f); err != nil {
					n.verticesMu[i].RUnlock()
					return nil, err
				}
			}

			metaCount := uint32(len(node.Metadata))
			if err := binary.Write(&buf, binary.BigEndian, metaCount); err != nil {
				n.verticesMu[i].RUnlock()
				return nil, err
			}
			for metaKey, metaVal := range node.Metadata {
				metaKeyBytes := []byte(metaKey)
				if len(metaKeyBytes) > 65535 {
					n.verticesMu[i].RUnlock()
					return nil, fmt.Errorf("metadata key too long: %s", metaKey)
				}
				if err := binary.Write(&buf, binary.BigEndian, uint16(len(metaKeyBytes))); err != nil {
					n.verticesMu[i].RUnlock()
					return nil, err
				}
				if _, err := buf.Write(metaKeyBytes); err != nil {
					n.verticesMu[i].RUnlock()
					return nil, err
				}
				switch v := metaVal.(type) {
				case int64:
					if err := buf.WriteByte(0); err != nil {
						n.verticesMu[i].RUnlock()
						return nil, err
					}
					if err := binary.Write(&buf, binary.BigEndian, v); err != nil {
						n.verticesMu[i].RUnlock()
						return nil, err
					}
				case string:
					if err := buf.WriteByte(1); err != nil {
						n.verticesMu[i].RUnlock()
						return nil, err
					}
					strBytes := []byte(v)
					if len(strBytes) > 65535 {
						n.verticesMu[i].RUnlock()
						return nil, fmt.Errorf("metadata string too long: %s", v)
					}
					if err := binary.Write(&buf, binary.BigEndian, uint16(len(strBytes))); err != nil {
						n.verticesMu[i].RUnlock()
						return nil, err
					}
					if _, err := buf.Write(strBytes); err != nil {
						n.verticesMu[i].RUnlock()
						return nil, err
					}
				default:
					n.verticesMu[i].RUnlock()
					return nil, fmt.Errorf("unsupported metadata type: %T", v)
				}
			}
		}
		n.verticesMu[i].RUnlock()
	}
	return buf.Bytes(), nil
}

func (n *noneVecSpace) LoadVertex(data []byte) error {
	buf := bytes.NewReader(data)
	var shards [EDGE_MAP_SHARD_COUNT]map[uint64]ENode

	for i := 0; i < EDGE_MAP_SHARD_COUNT; i++ {
		var count uint64
		if err := binary.Read(buf, binary.BigEndian, &count); err != nil {
			return err
		}
		m := make(map[uint64]ENode, count)
		for j := uint64(0); j < count; j++ {
			var key uint64
			if err := binary.Read(buf, binary.BigEndian, &key); err != nil {
				return err
			}

			var node ENode

			var vecLen uint32
			if err := binary.Read(buf, binary.BigEndian, &vecLen); err != nil {
				return err
			}
			node.Vector = make([]float32, vecLen)
			for d := 0; d < int(vecLen); d++ {
				if err := binary.Read(buf, binary.BigEndian, &node.Vector[d]); err != nil {
					return err
				}
			}

			var metaCount uint32
			if err := binary.Read(buf, binary.BigEndian, &metaCount); err != nil {
				return err
			}
			node.Metadata = make(map[string]any, metaCount)
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
				case 0:
					var val int64
					if err := binary.Read(buf, binary.BigEndian, &val); err != nil {
						return err
					}
					node.Metadata[metaKey] = val
				case 1:
					var strLen uint16
					if err := binary.Read(buf, binary.BigEndian, &strLen); err != nil {
						return err
					}
					strBytes := make([]byte, strLen)
					if _, err := io.ReadFull(buf, strBytes); err != nil {
						return err
					}
					node.Metadata[metaKey] = string(strBytes)
				default:
					return fmt.Errorf("unsupported metadata type tag: %d", typ)
				}
			}
			m[key] = node
		}
		shards[i] = m
	}
	for i := 0; i < EDGE_MAP_SHARD_COUNT; i++ {
		n.vertices[i] = shards[i]
		n.verticesMu[i] = &sync.RWMutex{}
	}
	return nil
}
