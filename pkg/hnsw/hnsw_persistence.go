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

package hnsw

import (
	"encoding/binary"
	"errors"
	"io"
	"sync/atomic"
	"unsafe"

	"github.com/google/uuid"
	"github.com/sjy-dv/nnv/pkg/gomath"
	"github.com/sjy-dv/nnv/pkg/hnsw/space"
)

var (
	InvalidSpaceTypeErr error = errors.New("Invalid space type")
	NoEntrypointErr     error = errors.New("No entrypoint")
)

func spaceToSpaceIdx(s space.Space) uint8 {
	switch s.(type) {
	case *space.Euclidean:
		return 1
	case *space.Manhattan:
		return 2
	case *space.Cosine:
		return 3
	}
	return 0
}

func spaceIdxToSpace(spaceIdx uint8) (space.Space, error) {
	switch spaceIdx {
	case 1:
		return space.NewEuclidean(), nil
	case 2:
		return space.NewManhattan(), nil
	case 3:
		return space.NewCosine(), nil
	}
	return nil, InvalidSpaceTypeErr
}

func (this *Hnsw) Save(w io.Writer, header bool) error {
	if header {
		if err := this.config.save(w); err != nil {
			return err
		}
		if err := binary.Write(w, binary.BigEndian, uint32(this.size)); err != nil {
			return err
		}
		if err := binary.Write(w, binary.BigEndian, spaceToSpaceIdx(this.space)); err != nil {
			return err
		}
	}

	if this.Len() == 0 {
		return nil
	}

	// Store entrypoint
	entrypoint := (*hnswVertex)(atomic.LoadPointer(&this.entrypoint))
	if entrypoint == nil {
		return NoEntrypointErr
	}
	if _, err := w.Write(entrypoint.id[:]); err != nil {
		return err
	}

	// Store vertices
	for _, verticesShard := range this.vertices {
		if err := binary.Write(w, binary.BigEndian, uint32(len(verticesShard))); err != nil {
			return err
		}

		for _, vertex := range verticesShard {
			if _, err := w.Write(vertex.id[:]); err != nil {
				return err
			}
			if err := binary.Write(w, binary.BigEndian, int32(vertex.level)); err != nil {
				return err
			}
			if err := vertex.vector.Save(w); err != nil {
				return err
			}
			if err := vertex.metadata.save(w); err != nil {
				return err
			}
		}
	}

	// Store edges
	for _, verticesShard := range this.vertices {
		for _, vertex := range verticesShard {
			if _, err := w.Write(vertex.id[:]); err != nil {
				return err
			}

			for l := vertex.level; l >= 0; l-- {
				edgesCount := 0
				for neighbor, _ := range vertex.edges[l] {
					if atomic.LoadUint32(&neighbor.deleted) == 0 {
						edgesCount++
					}
				}
				if err := binary.Write(w, binary.BigEndian, uint32(edgesCount)); err != nil {
					return err
				}
				for neighbor, distance := range vertex.edges[l] {
					if atomic.LoadUint32(&neighbor.deleted) == 1 {
						continue
					}
					if _, err := w.Write(neighbor.id[:]); err != nil {
						return err
					}
					if err := binary.Write(w, binary.BigEndian, distance); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func (this *Hnsw) Load(r io.Reader, header bool) error {
	if header {
		var size uint32
		var spaceIdx uint8
		if err := this.config.load(r); err != nil {
			return err
		}
		if err := binary.Read(r, binary.BigEndian, &size); err != nil {
			return err
		}
		if err := binary.Read(r, binary.BigEndian, &spaceIdx); err != nil {
			return err
		}
		space, err := spaceIdxToSpace(spaceIdx)
		if err != nil {
			return err
		}
		this.size = uint(size)
		this.space = space
	}

	var level int32
	var numEdges uint32
	var distance float32

	uuidBuf := make([]byte, 16)
	if _, err := r.Read(uuidBuf); err != nil {
		return err
	}
	entrypointId, err := uuid.FromBytes(uuidBuf)
	if err != nil {
		return err
	}

	this.len = 0
	// Load vertices
	var shardSize uint32
	var vertex *hnswVertex
	for i, _ := range this.vertices {
		if err := binary.Read(r, binary.BigEndian, &shardSize); err != nil {
			return err
		}
		this.len += uint64(shardSize)

		this.vertices[i] = make(map[uuid.UUID]*hnswVertex, int(shardSize))
		verticesShard := this.vertices[i]

		for i := 0; i < int(shardSize); i++ {
			if _, err := r.Read(uuidBuf); err != nil {
				return err
			}
			id, err := uuid.FromBytes(uuidBuf)
			if err != nil {
				return err
			}
			if err := binary.Read(r, binary.BigEndian, &level); err != nil {
				return err
			}

			vector := make(gomath.Vector, this.size)
			if err := vector.Load(r); err != nil {
				return err
			}

			metadata := make(Metadata)
			if err := metadata.load(r); err != nil {
				return err
			}

			vertex = newHnswVertex(id, vector, metadata, int(level))
			this.bytesSize += vertex.bytesSize()
			verticesShard[id] = vertex
		}
	}

	// Set entrypoint
	s, _ := this.getVerticesShard(entrypointId)
	atomic.StorePointer(&this.entrypoint, unsafe.Pointer(s[entrypointId]))

	// Load edges
	for _, verticesShard := range this.vertices {
		for i := 0; i < len(verticesShard); i++ {
			if _, err := r.Read(uuidBuf); err != nil {
				return err
			}
			id, err := uuid.FromBytes(uuidBuf)
			if err != nil {
				return err
			}

			vertex = verticesShard[id]
			for l := vertex.level; l >= 0; l-- {
				if err := binary.Read(r, binary.BigEndian, &numEdges); err != nil {
					return err
				}
				for j := 0; j < int(numEdges); j++ {
					if _, err := r.Read(uuidBuf); err != nil {
						return err
					}
					neighborId, err := uuid.FromBytes(uuidBuf)
					if err != nil {
						return err
					}
					if err := binary.Read(r, binary.BigEndian, &distance); err != nil {
						return err
					}
					s, _ = this.getVerticesShard(neighborId)
					vertex.edges[l][s[neighborId]] = distance
				}
			}
		}
	}

	return nil
}
