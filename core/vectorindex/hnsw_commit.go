package vectorindex

import (
	"encoding/binary"
	"errors"
	"io"
	"sync/atomic"
	"unsafe"

	"github.com/sjy-dv/nnv/edge"
	"github.com/sjy-dv/nnv/pkg/distance"
)

var (
	InvalidSpaceTypeErr error = errors.New("Invalid space type")
	NoEntrypointErr     error = errors.New("No entrypoint")
)

func distToDistIdx(dist distance.Space) uint8 {
	switch dist.Type() {
	case "cosine-dot":
		return 1
	case "l2-squared":
		return 2
	}
	return 0
}

func distIdxToDist(distIdx uint8) (distance.Space, error) {
	switch distIdx {
	case 1:
		return distance.NewCosine(), nil
	case 2:
		return distance.NewEuclidean(), nil
	}
	return nil, InvalidSpaceTypeErr
}

func idToBytes(id uint64) ([]byte, error) {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, id)
	return buf, nil
}

func bytesToId(byt []byte) (uint64, error) {
	if len(byt) != 8 {
		return 0, errors.New("invalid byte length for uint64")
	}
	return binary.BigEndian.Uint64(byt), nil
}

func (xx *Hnsw) Commit(w io.Writer, header bool) error {
	if header {
		if err := xx.config.save(w); err != nil {
			return err
		}
		if err := binary.Write(w, binary.BigEndian, uint32(xx.dim)); err != nil {
			return err
		}
		if err := binary.Write(w, binary.BigEndian, distToDistIdx(xx.distancer)); err != nil {
			return err
		}
	}

	if xx.Len() == 0 {
		return nil
	}

	entrypoint := (*hnswVertex)(atomic.LoadPointer(&xx.entrypoint))
	if entrypoint == nil {
		return NoEntrypointErr
	}
	ebid, err := idToBytes(entrypoint.id)
	if err != nil {
		return err
	}
	if _, err := w.Write(ebid); err != nil {
		return err
	}

	for _, verticShard := range xx.vertices {
		if err := binary.Write(w, binary.BigEndian, uint32(len(verticShard))); err != nil {
			return err
		}

		for _, vertex := range verticShard {
			byid, err := idToBytes(vertex.id)
			if err != nil {
				return err
			}
			if _, err := w.Write(byid); err != nil {
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

	for _, verticesShard := range xx.vertices {
		for _, vertex := range verticesShard {
			byid, err := idToBytes(vertex.id)
			if err != nil {
				return err
			}
			if _, err := w.Write(byid); err != nil {
				return err
			}

			for l := vertex.level; l >= 0; l-- {
				edgesCount := 0
				for neighbor := range vertex.edges[l] {
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
					byid, err := idToBytes(neighbor.id)
					if err != nil {
						return err
					}
					if _, err := w.Write(byid); err != nil {
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

func (xx *Hnsw) Load(r io.Reader, header bool) error {
	if header {
		var size uint32
		var distIdx uint8
		if err := xx.config.load(r); err != nil {
			return err
		}
		if err := binary.Read(r, binary.BigEndian, &size); err != nil {
			return err
		}
		if err := binary.Read(r, binary.BigEndian, &distIdx); err != nil {
			return err
		}
		dist, err := distIdxToDist(distIdx)
		if err != nil {
			return err
		}
		xx.dim = uint(size)
		xx.distancer = dist
	}

	var level int32
	var numEdges uint32
	var distance float32

	idBuf := make([]byte, 8)
	if _, err := r.Read(idBuf); err != nil {
		return err
	}
	entrypointId, err := bytesToId(idBuf)
	if err != nil {
		return err
	}

	xx.len = 0
	// Load vertices
	var shardSize uint32
	var vertex *hnswVertex
	for i := range xx.vertices {
		if err := binary.Read(r, binary.BigEndian, &shardSize); err != nil {
			return err
		}
		xx.len += uint64(shardSize)

		xx.vertices[i] = make(map[uint64]*hnswVertex, int(shardSize))
		verticesShard := xx.vertices[i]

		for i := 0; i < int(shardSize); i++ {
			if _, err := r.Read(idBuf); err != nil {
				return err
			}
			id, err := bytesToId(idBuf)
			if err != nil {
				return err
			}
			if err := binary.Read(r, binary.BigEndian, &level); err != nil {
				return err
			}

			vector := make(edge.Vector, xx.dim)
			if err := vector.Load(r); err != nil {
				return err
			}

			metadata := make(Metadata)
			if err := metadata.load(r); err != nil {
				return err
			}

			vertex = newHnswVertex(id, vector, metadata, int(level))
			xx.bytesSize += vertex.bytesSize()
			verticesShard[id] = vertex
		}
	}

	// Set entrypoint
	s, _ := xx.getVerticesShard(entrypointId)
	atomic.StorePointer(&xx.entrypoint, unsafe.Pointer(s[entrypointId]))

	// Load edges
	for _, verticesShard := range xx.vertices {
		for i := 0; i < len(verticesShard); i++ {
			if _, err := r.Read(idBuf); err != nil {
				return err
			}
			id, err := bytesToId(idBuf)
			if err != nil {
				return err
			}

			vertex = verticesShard[id]
			for l := vertex.level; l >= 0; l-- {
				if err := binary.Read(r, binary.BigEndian, &numEdges); err != nil {
					return err
				}
				for j := 0; j < int(numEdges); j++ {
					if _, err := r.Read(idBuf); err != nil {
						return err
					}
					neighborId, err := bytesToId(idBuf)
					if err != nil {
						return err
					}
					if err := binary.Read(r, binary.BigEndian, &distance); err != nil {
						return err
					}
					s, _ = xx.getVerticesShard(neighborId)
					vertex.edges[l][s[neighborId]] = distance
				}
			}
		}
	}

	return nil
}
