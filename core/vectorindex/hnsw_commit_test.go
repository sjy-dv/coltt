package vectorindex

import (
	"bytes"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/sjy-dv/nnv/pkg/distance"
	"github.com/sjy-dv/nnv/pkg/gomath"
	"github.com/stretchr/testify/assert"
)

func hnswIsEqual(a, b *Hnsw) error {
	if a.Len() != b.Len() {
		return errors.New("Length missmatch")
	}
	fmt.Println(a.bytesSize, b.bytesSize)
	// if a.BytesSize() != b.BytesSize() {
	// 	return errors.New("Bytes size missmatch")
	// }

	for i, shard := range a.vertices {
		otherShard := b.vertices[i]
		if len(shard) != len(otherShard) {
			return errors.New("Not same shard len")
		}
		for _, vertex := range shard {
			otherVertex, exists := otherShard[vertex.id]
			if !exists {
				return errors.New("Other vertex does not exist")
			}
			if vertex.level != otherVertex.level {
				return errors.New("Other vertex level does not match")
			}
			if len(vertex.vector) != len(otherVertex.vector) {
				return errors.New("Other vertex vector size does not match")
			}
			for j := 0; j < len(vertex.vector); j++ {
				if gomath.Abs(vertex.vector[j]-otherVertex.vector[j]) > 1e-4 {
					return errors.New("Other vector value does not match")
				}
			}
			if len(vertex.metadata) != len(otherVertex.metadata) {
				return errors.New("Metadata size does not match")
			}
			if vertex.bytesSize() != otherVertex.bytesSize() {
				return errors.New("Vertex bytes size missmatch")
			}
			for k, v := range vertex.metadata {
				if otherVertex.metadata[k] != v {
					return errors.New("Metadata value missmatch")
				}
			}
			for l := vertex.level; l >= 0; l-- {
				neighborDistances := make(map[uint64]float32)
				otherNeighborDistances := make(map[uint64]float32)
				for neighbor, distance := range vertex.edges[l] {
					if atomic.LoadUint32(&neighbor.deleted) == 0 {
						neighborDistances[neighbor.id] = distance
					}
				}
				for neighbor, distance := range otherVertex.edges[l] {
					if atomic.LoadUint32(&neighbor.deleted) == 0 {
						otherNeighborDistances[neighbor.id] = distance
					}
				}
				if len(neighborDistances) != len(otherNeighborDistances) {
					return errors.New("Edges count not the same")
				}
				for id, distance := range neighborDistances {
					otherDistance, exists := otherNeighborDistances[id]
					if !exists {
						return errors.New("Other neighbor does not exist")
					}
					if gomath.Abs(distance-otherDistance) > 1e-4 {
						return errors.New("Distances do not match")
					}
				}
			}
		}
	}
	return nil
}

func generateRandomIndex(dim, size int, dist distance.Space) *Hnsw {
	insertKeys := make(map[uint64]struct{})

	index := NewHnsw(uint(dim), dist)
	delOffset := int(size / 10)
	for i := 0; i < size; i++ {
		if i > delOffset && (gomath.RandomUniform() <= 0.2) {
			var key uint64
			for k := range insertKeys {
				key = k
				break
			}
			delete(insertKeys, key)
			index.Remove(key)
		} else {
			id := uint64(i)
			insertKeys[id] = struct{}{}
			index.Insert(id, gomath.RandomUniformVector(dim), Metadata{"foo": fmt.Sprintf("bar: %d", i), "id": fmt.Sprintf("%d", id)}, index.RandomLevel())
		}
	}
	return index
}

func TestHnswCommitAndLoad(t *testing.T) {
	l2index := generateRandomIndex(128, 1000, distance.NewEuclidean())
	cosindex := generateRandomIndex(128, 1000, distance.NewCosine())

	var l2buf bytes.Buffer
	var cosbuf bytes.Buffer

	err := l2index.Commit(&l2buf, true)
	assert.Nil(t, err)
	if err != nil {
		return
	}
	err = cosindex.Commit(&cosbuf, true)
	assert.Nil(t, err)
	if err != nil {
		return
	}

	copyl2 := NewHnsw(128, distance.NewEuclidean())
	copycos := NewHnsw(128, distance.NewCosine())

	err = copyl2.Load(&l2buf, true)
	assert.Nil(t, err)
	if err != nil {
		return
	}
	err = copycos.Load(&cosbuf, true)
	assert.Nil(t, err)
	if err != nil {
		return
	}

	assert.Nil(t, hnswIsEqual(l2index, copyl2))
	assert.Nil(t, hnswIsEqual(cosindex, copycos))
}

func TestHnswSimple(t *testing.T) {
	index := generateRandomIndex(128, 1000, distance.NewEuclidean())

	var buf bytes.Buffer
	err := index.Commit(&buf, true)
	assert.Nil(t, err)
	if err != nil {
		return
	}

	otherIndex := NewHnsw(128, distance.NewEuclidean())
	err = otherIndex.Load(&buf, true)
	assert.Nil(t, err)
	if err != nil {
		return
	}

	assert.Nil(t, hnswIsEqual(index, otherIndex))
}
