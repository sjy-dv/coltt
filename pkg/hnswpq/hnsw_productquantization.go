package hnswpq

import (
	"container/heap"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sync"

	"github.com/sjy-dv/nnv/pkg/distancepq"
	"github.com/sjy-dv/nnv/pkg/gomath"
	"github.com/sjy-dv/nnv/pkg/models"
	"github.com/sjy-dv/nnv/pkg/queue"
)

type HnswPQs struct {
	Collections map[string]*Hnsw
	gLock       sync.RWMutex
}

func NewProductQuantizationHnsw() *HnswPQs {
	return &HnswPQs{
		Collections: make(map[string]*Hnsw),
	}
}

func (xx *HnswPQs) CreateCollection(collectionName string, config models.HnswConfig, params models.ProductQuantizerParameters) error {
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
		EmptyNodes:     make([]uint64, 0),
		CollectionName: collectionName,
		PQ:             pq,
	}
	xx.gLock.Unlock()

	return nil
}

func (xx *HnswPQs) Genesis(collectionName string, config models.HnswConfig) {
	dummyVector := make(gomath.Vector, config.Dim)
	genesisNode := Node{
		Id:        0,
		Layer:     0,
		Vectors:   dummyVector,
		LinkNodes: make([][]uint64, config.Mmax0+1),
		Centroids: xx.Collections[collectionName].PQ.encode(dummyVector),
	}
	xx.Collections[collectionName].NodeList.lock.Lock()
	xx.Collections[collectionName].NodeList.Nodes[0] = genesisNode
	xx.Collections[collectionName].NodeList.lock.Unlock()
}

func (xx *HnswPQs) DropCollection(collectionName string) error {
	xx.gLock.Lock()
	//safe memory & gc
	xx.Collections[collectionName] = nil
	delete(xx.Collections, collectionName)
	xx.gLock.Unlock()
	return nil
}

// when commitId is zero, reuse empty node space
func (xx *HnswPQs) Insert(collectionName string, commitID uint64, vec gomath.Vector) error {

	centroidIds := xx.Collections[collectionName].PQ.encode(vec)
	node := Node{
		Vectors:   vec,
		Layer:     int(math.Floor(-math.Log(rand.Float64()) * xx.Collections[collectionName].Ml)),
		LinkNodes: make([][]uint64, xx.Collections[collectionName].M+1),
		IsEmpty:   false,
		Centroids: centroidIds,
	}

	var nodeId uint64
	nodeId = commitID
	node.Id = nodeId
	xx.Collections[collectionName].NodeList.Nodes = append(xx.Collections[collectionName].NodeList.Nodes, node)

	xx.Collections[collectionName].NodeList.lock.Unlock()
	_, err := xx.Collections[collectionName].PQ.Set(nodeId, vec)
	if err != nil {
		return fmt.Errorf(pointPQSetErr, err)
	}

	_, err = xx.Collections[collectionName].PQ.Get(nodeId)
	if err != nil {
		return fmt.Errorf(pointPQGetErr, err)
	}
	xx.Collections[collectionName].PQ.Dirty(nodeId)

	curObj := &xx.Collections[collectionName].NodeList.Nodes[xx.Collections[collectionName].Ep]
	curDist := xx.Collections[collectionName].PQ.DistanceFromCentroidIDs(vec, curObj.Centroids)

	heapCandidates := &queue.PriorityQueue{Order: false, Items: make([]*queue.Item, 0)}
	heap.Init(heapCandidates)
	heap.Push(heapCandidates, &queue.Item{Distance: curDist, NodeID: curObj.Id})

	for level := min(int(node.Layer), int(xx.Collections[collectionName].MaxLevel)); level >= 0; level-- {
		err := xx.Collections[collectionName].searchLayer(
			vec,
			&queue.Item{Distance: curDist, NodeID: curObj.Id},
			heapCandidates,
			int(xx.Collections[collectionName].Efconstruction),
			uint(level),
		)
		if err != nil {
			return err
		}

		switch xx.Collections[collectionName].Heuristic {
		case false:
			xx.Collections[collectionName].SelectNeighboursSimple(heapCandidates, int(xx.Collections[collectionName].M))
		case true:
			xx.Collections[collectionName].SelectNeighboursHeuristic(heapCandidates, int(xx.Collections[collectionName].M), false)
		}
		node.LinkNodes[level] = make([]uint64, heapCandidates.Len())
		for i := heapCandidates.Len() - 1; i >= 0; i-- {
			candidate := heap.Pop(heapCandidates).(*queue.Item)
			node.LinkNodes[level][i] = candidate.NodeID
		}
		xx.Collections[collectionName].NodeList.lock.Lock()
		xx.Collections[collectionName].NodeList.Nodes[nodeId].LinkNodes = node.LinkNodes
		xx.Collections[collectionName].NodeList.lock.Unlock()

		xx.Collections[collectionName].NodeList.lock.Lock()
		for _, neighbourNode := range node.LinkNodes[level] {
			xx.Collections[collectionName].addConnections(neighbourNode, nodeId, level)
		}
		xx.Collections[collectionName].NodeList.lock.Unlock()
	}

	if node.Layer > xx.Collections[collectionName].MaxLevel {
		xx.Collections[collectionName].hlock.Lock()
		xx.Collections[collectionName].Ep = int64(nodeId)
		xx.Collections[collectionName].MaxLevel = node.Layer
		xx.Collections[collectionName].hlock.Unlock()
	}
	return nil
}

func (xx *HnswPQs) Search(collectionName string, vec []float32, topCandidates *queue.PriorityQueue, K int, efSearch int) error {

	pq := xx.Collections[collectionName].PQ

	distFn := func(q []float32, centroids []uint8) float32 {
		return pq.DistanceFromCentroidIDs(q, centroids)
	}

	curObj := &xx.Collections[collectionName].NodeList.Nodes[xx.Collections[collectionName].Ep]
	curDist := distFn(vec, curObj.Centroids)

	heapCandidates := &queue.PriorityQueue{Order: false, Items: []*queue.Item{}}
	heap.Init(heapCandidates)
	heap.Push(heapCandidates, &queue.Item{Distance: curDist, NodeID: curObj.Id})

	for level := int(xx.Collections[collectionName].MaxLevel); level >= 0; level-- {
		err := xx.Collections[collectionName].searchLayer(
			vec,
			&queue.Item{Distance: curDist, NodeID: curObj.Id},
			heapCandidates,
			efSearch,
			uint(level),
		)
		if err != nil {
			return err
		}

		switch xx.Collections[collectionName].Heuristic {
		case false:
			xx.Collections[collectionName].SelectNeighboursSimple(heapCandidates, int(xx.Collections[collectionName].M))
		case true:
			xx.Collections[collectionName].SelectNeighboursHeuristic(heapCandidates, int(xx.Collections[collectionName].M), false)
		}
	}

	for heapCandidates.Len() > K {
		_ = heap.Pop(heapCandidates).(*queue.Item)
	}

	for _, item := range heapCandidates.Items {
		heap.Push(topCandidates, item)
	}

	return nil
}

func (xx *HnswPQs) FailAppointNode(collectioName string, failID uint64) error {
	xx.Collections[collectioName].hlock.Lock()
	defer xx.Collections[collectioName].hlock.Unlock()
	xx.Collections[collectioName].EmptyNodes = append(xx.Collections[collectioName].EmptyNodes, failID)
	return nil
}

func (xx *HnswPQs) AppointNode(collectionName string) uint64 {
	if xx.IsEmpty(collectionName) {
		return 0
	}
	xx.Collections[collectionName].hlock.Lock()
	defer xx.Collections[collectionName].hlock.Unlock()
	lastIndex := len(xx.Collections[collectionName].EmptyNodes) - 1
	last := xx.Collections[collectionName].EmptyNodes[lastIndex]
	xx.Collections[collectionName].EmptyNodes = xx.Collections[collectionName].EmptyNodes[:lastIndex]
	return last
}

func (xx *HnswPQs) IsEmpty(collectionName string) bool {
	xx.Collections[collectionName].hlock.RLock()
	defer xx.Collections[collectionName].hlock.RUnlock()
	return len(xx.Collections[collectionName].EmptyNodes) == 0
}

func (xx *HnswPQs) Delete(collectioName string, nodeId uint64) error {
	err := xx.Collections[collectioName].removeConnection(nodeId)
	return err
}

// remove => empty ID
// new Fit => ?.? not
func (xx *Hnsw) removeConnection(nodeId uint64) error {
	node := &xx.NodeList.Nodes[nodeId]
	if node.Id == 0 && !node.IsEmpty {
		return errors.New("node not found")
	}

	for level := 0; level <= xx.MaxLevel; level++ {
		xx.NodeList.lock.Lock()
		connections := node.LinkNodes[level]
		for _, neighbourId := range connections {
			neighbor := &xx.NodeList.Nodes[neighbourId]
			newLinks := []uint64{}
			for _, link := range neighbor.LinkNodes[level] {
				if link != nodeId {
					newLinks = append(newLinks, link)
				}
			}
			neighbor.LinkNodes[level] = newLinks
		}
		xx.NodeList.lock.Unlock()
	}

	xx.NodeList.lock.Lock()
	xx.NodeList.Nodes[nodeId] = Node{
		Id:      nodeId,
		IsEmpty: true,
	}
	xx.NodeList.lock.Unlock()
	xx.hlock.Lock()
	xx.EmptyNodes = append(xx.EmptyNodes, nodeId)
	xx.hlock.Unlock()
	return nil
}
