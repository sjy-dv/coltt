package hnswpq

import (
	"fmt"
	"math"
	"math/rand"
	"sync"

	"github.com/sjy-dv/nnv/pkg/distancepq"
	"github.com/sjy-dv/nnv/pkg/gomath"
	"github.com/sjy-dv/nnv/pkg/hnsw"
)

type HnswPQs struct {
	Collections       map[string]*Hnsw
	gLock             sync.RWMutex
	ProductQuantizers map[string]*productQuantizer
}

func (xx *HnswPQs) NewProductQuantizationHnsw() *HnswPQs {
	return &HnswPQs{
		Collections:       make(map[string]*Hnsw),
		ProductQuantizers: make(map[string]*productQuantizer),
	}
}

func (xx *HnswPQs) CreateCollection(genesisId uint64, collectionName string, config hnsw.HnswConfig, params ProductQuantizerParameters) error {
	//[exists collection] already check in highmem <-

	pq, err := newProductQuantizer(config.DistanceType, params, int(config.Dim))
	if err != nil {
		return err
	}
	xx.gLock.Lock()

	xx.Collections[collectionName] = &Hnsw{
		Efconstruction: config.Efconstruction,
		M:              config.M,
		Mmax:           config.Mmax,
		Mmax0:          config.Mmax0,
		Ml:             config.Ml,
		Ep:             config.Ep,
		MaxLevel:       config.MaxLevel,
		Dim:            config.Dim,
		Heuristic:      config.Heuristic,
		DistFn:         distancepq.GetFloatDistanceFn(config.DistanceType),
		DistFnName:     config.DistanceType,
		NodeList:       NodeList{Nodes: make([]Node, 1)},
		CollectionName: collectionName,
		EmptyNodes:     make([]uint64, 0),
	}
	xx.ProductQuantizers[collectionName] = new(productQuantizer)
	xx.ProductQuantizers[collectionName] = pq
	xx.gLock.Unlock()
	genesisNode := Node{
		Id:        0,
		Layer:     0,
		Vectors:   make(gomath.Vector, config.Dim),
		LinkNodes: make([][]uint64, config.Mmax0+1),
	}
	xx.Collections[collectionName].NodeList.lock.Lock()
	xx.Collections[collectionName].NodeList.Nodes[0] = genesisNode
	xx.Collections[collectionName].NodeList.lock.Unlock()

	return nil
}

func (xx *HnswPQs) DropCollection(collectionName string) error {
	xx.gLock.Lock()
	//safe memory & gc
	xx.Collections[collectionName] = nil
	xx.ProductQuantizers[collectionName] = nil
	delete(xx.Collections, collectionName)
	delete(xx.ProductQuantizers, collectionName)
	xx.gLock.Unlock()
	return nil
}

// when commitId is zero, reuse empty node space
func (xx *HnswPQs) Insert(collectionName string, commitID uint64, vec gomath.Vector) error {

	centroidIds := xx.ProductQuantizers[collectionName].encode(vec)

	node := Node{
		Vectors:   vec,
		Layer:     int(math.Floor(-math.Log(rand.Float64()) * xx.Collections[collectionName].Ml)),
		LinkNodes: make([][]uint64, xx.Collections[collectionName].M+1),
		IsEmpty:   false,
		Centroids: centroidIds,
	}

	var nodeId uint64
	xx.Collections[collectionName].NodeList.lock.Lock()
	if commitID == 0 {
		nodeId = xx.Collections[collectionName].EmptyNodes[len(xx.Collections)-1]
		node.Id = nodeId
		xx.Collections[collectionName].EmptyNodes = xx.Collections[collectionName].EmptyNodes[:len(xx.Collections[collectionName].EmptyNodes)-1]
		xx.Collections[collectionName].NodeList.Nodes[nodeId] = node
	} else {
		nodeId = commitID
		node.Id = nodeId
		xx.Collections[collectionName].NodeList.Nodes = append(xx.Collections[collectionName].NodeList.Nodes, node)
	}

	xx.Collections[collectionName].NodeList.lock.Unlock()
	_, err := xx.ProductQuantizers[collectionName].Set(nodeId, vec)
	if err != nil {
		return fmt.Errorf(pointPQSetErr, err)
	}

	_, err = xx.ProductQuantizers[collectionName].Get(nodeId)
	if err != nil {
		return fmt.Errorf(pointPQGetErr, err)
	}
	xx.ProductQuantizers[collectionName].Dirty(nodeId)
	
	curObj := &xx.Collections[collectionName].NodeList.Nodes[xx.Collections[collectionName].Ep]
	curDist := 
}
