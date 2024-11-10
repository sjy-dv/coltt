package edge

import (
	"fmt"
	"sync"

	"github.com/sjy-dv/nnv/pkg/distance"
	"github.com/sjy-dv/nnv/pkg/gomath"
)

type EdgeVector struct {
	dimension      int
	vectors        map[uint64]gomath.Vector
	collectionName string
	distance       distance.Space
	lock           sync.RWMutex
}

type EdgeVectors struct {
	Edges map[string]*EdgeVector
	lock  sync.RWMutex
}

func NewEdgeVectorCollection() *EdgeVectors {
	return &EdgeVectors{
		Edges: make(map[string]*EdgeVector),
	}
}

type CollectionConfig struct {
	dimension      int
	collectionName string
	distance       distance.Space
}

func (xx *EdgeVectors) CreateCollection(config CollectionConfig) error {
	xx.lock.RLock()
	_, ok := xx.Edges[config.collectionName]
	xx.lock.RUnlock()
	if ok {
		return fmt.Errorf(ErrCollectionExists, config.collectionName)
	}
	xx.lock.Lock()
	xx.Edges[config.collectionName].collectionName = config.collectionName
	xx.Edges[config.collectionName].dimension = config.dimension
	xx.Edges[config.collectionName].distance = config.distance
	xx.Edges[config.collectionName].vectors = make(map[uint64]gomath.Vector)
	xx.lock.Unlock()
	return nil
}

// add file logic
func (xx *EdgeVectors) DropCollection(collectionName string) error {
	xx.lock.RLock()
	_, ok := xx.Edges[collectionName]
	xx.lock.RUnlock()
	if !ok {
		return nil
	}
	xx.lock.Lock()
	delete(xx.Edges, collectionName)
	xx.lock.Unlock()
	return nil
}

func (xx *EdgeVectors) InsertVector(collectionName string, commitId uint64, vector gomath.Vector) error {
	xx.lock.RLock()
	basis, ok := xx.Edges[collectionName]
	xx.lock.RUnlock()
	if !ok {
		return fmt.Errorf(ErrCollectionNotFound, collectionName)
	}
	basis.lock.Lock()
	basis.vectors[commitId] = vector
	basis.lock.Unlock()
	return nil
}

func (xx *EdgeVectors) UpdateVector(collectionName string, id uint64, vector gomath.Vector) error {
	xx.lock.RLock()
	basis, ok := xx.Edges[collectionName]
	xx.lock.RUnlock()
	if !ok {
		return fmt.Errorf(ErrCollectionNotFound, collectionName)
	}
	basis.lock.Lock()
	basis.vectors[id] = vector
	basis.lock.Unlock()
	return nil
}

func (xx *EdgeVectors) RemoveVector(collectionName string, id uint64) error {
	xx.lock.RLock()
	basis, ok := xx.Edges[collectionName]
	xx.lock.RUnlock()
	if !ok {
		return fmt.Errorf(ErrCollectionNotFound, collectionName)
	}
	basis.lock.Lock()
	delete(basis.vectors, id)
	basis.lock.Unlock()
	return nil
}

func (xx *EdgeVectors) FullScan(collectionName string, target gomath.Vector, topK int) (*ResultSet, error) {
	rs := NewResultSet(topK)
	xx.lock.RLock()
	basis, ok := xx.Edges[collectionName]
	xx.lock.RUnlock()
	if !ok {
		return nil, fmt.Errorf(ErrCollectionNotFound, collectionName)
	}
	basis.lock.RLock()
	defer basis.lock.RUnlock()
	for index, vector := range basis.vectors {
		sim := basis.distance.Distance(target, vector)
		rs.AddResult(ID(index), sim)
	}
	return rs, nil
}
