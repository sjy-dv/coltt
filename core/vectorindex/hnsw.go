package vectorindex

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/sjy-dv/nnv/edge"
	"github.com/sjy-dv/nnv/pkg/distance"
	"github.com/sjy-dv/nnv/pkg/gomath"
	"github.com/sjy-dv/nnv/pkg/sharding"
)

const VERTICES_MAP_SHARD_COUNT int = 16

var (
	ItemNotFoundError      error = errors.New("Item not found")
	ItemAlreadyExistsError error = errors.New("Item already exists")
)

type Hnsw struct {
	dim       uint
	bytesSize uint64
	distancer distance.Space
	config    *hnswConfig

	len        uint64
	vertices   [VERTICES_MAP_SHARD_COUNT]map[uint64]*hnswVertex
	verticesMu [VERTICES_MAP_SHARD_COUNT]*sync.RWMutex

	entrypoint unsafe.Pointer
}

func NewHnsw(dim uint, distancer distance.Space, option ...HnswOption) *Hnsw {
	index := &Hnsw{
		dim:       dim,
		distancer: distancer,
		config:    newHnswConfig(option),

		len:        0,
		entrypoint: nil,
	}

	for i := 0; i < VERTICES_MAP_SHARD_COUNT; i++ {
		index.vertices[i] = make(map[uint64]*hnswVertex)
		index.verticesMu[i] = &sync.RWMutex{}
	}

	return index
}

func (xx *Hnsw) Info() string {
	return fmt.Sprintf("HNSW(dim: %d, distancer: %s, config={%s})", xx.dim, xx.distancer.Type(), xx.config)
}

func (xx *Hnsw) Dim() uint32 {
	return uint32(xx.dim)
}

func (xx *Hnsw) Len() int {
	return int(atomic.LoadUint64(&xx.len))
}

func (xx *Hnsw) Config() ProtoConfig {
	return ProtoConfig{
		SearchAlgorithm:           strings.ToLower(xx.config.searchAlgorithm.String()),
		LevelMultiplier:           xx.config.levelMultiplier,
		Ef:                        xx.config.ef,
		EfConstruction:            xx.config.efConstruction,
		M:                         xx.config.m,
		MMax:                      xx.config.mMax,
		MMax0:                     xx.config.mMax0,
		HeuristicExtendCandidates: xx.config.heuristicExtendCandidates,
		HeuristicKeepPruned:       xx.config.heuristicKeepPruned,
	}
}

func (xx *Hnsw) Distance() string {
	return xx.distancer.Type()
}

func (xx *Hnsw) Insert(id uint64, value edge.Vector, metadata Metadata, vertexLevel int) error {
	if xx.distancer.Type() == "cosine-dot" {
		value = Normalize(value)
	}
	var vertex *hnswVertex
	if (*hnswVertex)(atomic.LoadPointer(&xx.entrypoint)) == nil {
		vertex = newHnswVertex(id, value, metadata, 0)
		if err := xx.storeVertex(vertex); err != nil {
			return err
		}
		if atomic.CompareAndSwapPointer(&xx.entrypoint, nil, unsafe.Pointer(vertex)) {
			return nil
		} else {
			vertex.setLevel(vertexLevel)
		}
	} else {
		vertex = newHnswVertex(id, value, metadata, vertexLevel)
		if err := xx.storeVertex(vertex); err != nil {
			return err
		}
	}

	entrypoint := (*hnswVertex)(atomic.LoadPointer(&xx.entrypoint))
	minDistance := xx.distancer.Distance(vertex.vector, entrypoint.vector)
	for l := entrypoint.level; l > vertex.level; l-- {
		entrypoint, minDistance = xx.greedyClosestNeighbor(vertex.vector, entrypoint, minDistance, l)
	}

	for l := gomath.MinInt(entrypoint.level, vertex.level); l >= 0; l-- {
		neighbors := xx.searchLevel(vertex.vector, entrypoint, xx.config.efConstruction, l)

		switch xx.config.searchAlgorithm {
		case HnswSearchSimple:
			neighbors = xx.selectNeighbors(neighbors, xx.config.m)
		case HnswSearchHeuristic:
			neighbors = xx.selectNeighborsHeuristic(vertex.vector, neighbors, xx.config.m, l, xx.config.heuristicExtendCandidates, xx.config.heuristicKeepPruned)
		}

		mMax := xx.config.mMax
		if l == 0 {
			mMax = xx.config.mMax0
		}

		for neighbors.Len() > 0 {
			item := neighbors.Pop()
			neighbor := item.Value().(*hnswVertex)
			entrypoint = neighbor

			vertex.addEdge(l, neighbor, item.Priority())
			neighbor.addEdge(l, vertex, item.Priority())

			if neighbor.edgesCount(l) > mMax {
				xx.pruneNeighbors(neighbor, mMax, l)
			}
		}
	}

	entrypoint = (*hnswVertex)(atomic.LoadPointer(&xx.entrypoint))
	if entrypoint != nil && vertex.level > entrypoint.level {
		atomic.CompareAndSwapPointer(&xx.entrypoint, xx.entrypoint, unsafe.Pointer(vertex))
	}

	return nil
}

func (xx *Hnsw) Get(id uint64) (edge.Vector, error) {
	m, mu := xx.getVerticesShard(id)
	mu.RLock()
	defer mu.RUnlock()

	if vertex, exists := m[id]; exists {
		return vertex.vector, nil
	}
	return nil, ItemNotFoundError
}

func (xx *Hnsw) GetVertex(id uint64) (*hnswVertex, error) {
	m, mu := xx.getVerticesShard(id)
	mu.RLock()
	defer mu.RUnlock()

	if vertex, exists := m[id]; exists {
		return vertex, nil
	}
	return nil, ItemNotFoundError
}

func (xx *Hnsw) Remove(id uint64) error {
	vertex, err := xx.removeVertex(id)
	if err != nil {
		return err
	}

	currEntrypoint := atomic.LoadPointer(&xx.entrypoint)
	if (*hnswVertex)(currEntrypoint) == vertex {
		minDistance := gomath.MaxFloat
		var closestNeighbor *hnswVertex = nil

		for l := vertex.level; l >= 0; l-- {
			vertex.edgeMutexes[l].RLock()
			for neighbor, distance := range vertex.edges[l] {
				if distance < minDistance {
					minDistance = distance
					closestNeighbor = neighbor
				}
			}
			vertex.edgeMutexes[l].RUnlock()

			if closestNeighbor != nil {
				break
			}
		}
		atomic.CompareAndSwapPointer(&xx.entrypoint, currEntrypoint, unsafe.Pointer(closestNeighbor))
	}

	for l := vertex.level; l >= 0; l-- {
		mMax := xx.config.mMax
		if l == 0 {
			mMax = xx.config.mMax0
		}

		vertex.edgeMutexes[l].RLock()
		neighbors := make([]*hnswVertex, len(vertex.edges[l]))
		i := 0
		for neighbor, _ := range vertex.edges[l] {
			neighbors[i] = neighbor
			i++
		}
		vertex.edgeMutexes[l].RUnlock()

		for _, neighbor := range neighbors {
			neighbor.removeEdge(l, vertex)
			xx.pruneNeighbors(neighbor, mMax, l)
		}
	}

	return nil
}

func (xx *Hnsw) Search(ctx context.Context, query edge.Vector, k uint) (SearchResult, error) {
	if xx.distancer.Type() == "cosine-dot" {
		query = Normalize(query)
	}

	entrypoint := (*hnswVertex)(atomic.LoadPointer(&xx.entrypoint))
	if entrypoint == nil {
		return make(SearchResult, 0), nil
	}

	minDistance := xx.distancer.Distance(query, entrypoint.vector)
	for l := entrypoint.level; l > 0; l-- {
		entrypoint, minDistance = xx.greedyClosestNeighbor(query, entrypoint, minDistance, l)
	}

	ef := gomath.MaxInt(xx.config.ef, int(k))
	neighbors := xx.searchLevel(query, entrypoint, ef, 0)

	switch xx.config.searchAlgorithm {
	case HnswSearchSimple:
		neighbors = xx.selectNeighbors(neighbors, int(k))
	case HnswSearchHeuristic:
		neighbors = xx.selectNeighborsHeuristic(query, neighbors, int(k), 0, xx.config.heuristicExtendCandidates, xx.config.heuristicKeepPruned)
	}

	n := gomath.MinInt(int(k), neighbors.Len())
	result := make(SearchResult, n)
	for i := n - 1; i >= 0; i-- {
		item := neighbors.Pop()
		result[i].Id = item.Value().(*hnswVertex).Id()
		result[i].Metadata = item.Value().(*hnswVertex).Metadata()
		result[i].Score = item.Priority()
	}

	return result, nil
}

func (xx *Hnsw) RandomLevel() int {
	return gomath.Floor(gomath.RandomExponential(xx.config.levelMultiplier))
}

func (xx *Hnsw) getVerticesShard(id uint64) (map[uint64]*hnswVertex, *sync.RWMutex) {
	shardIdx := sharding.ShardVertex(id, uint64(VERTICES_MAP_SHARD_COUNT))
	return xx.vertices[shardIdx], xx.verticesMu[shardIdx]
}

func (xx *Hnsw) storeVertex(vertex *hnswVertex) error {
	m, mu := xx.getVerticesShard(vertex.id)
	defer mu.Unlock()
	mu.Lock()

	if _, exists := m[vertex.id]; exists {
		return ItemAlreadyExistsError
	}

	m[vertex.id] = vertex
	atomic.AddUint64(&xx.len, 1)
	atomic.AddUint64(&xx.bytesSize, 0)
	return nil
}

func (xx *Hnsw) removeVertex(id uint64) (*hnswVertex, error) {
	m, mu := xx.getVerticesShard(id)
	defer mu.Unlock()
	mu.Lock()

	if vertex, exists := m[id]; exists {
		delete(m, id)
		atomic.AddUint64(&xx.len, ^uint64(0))
		// atomic.AddUint64(&xx.bytesSize, ^uint64(vertex.bytesSize()-1))
		vertex.setDeleted()
		return vertex, nil
	}

	return nil, ItemNotFoundError
}

func (xx *Hnsw) greedyClosestNeighbor(query edge.Vector, entrypoint *hnswVertex, minDistance float32, level int) (*hnswVertex, float32) {
	for {
		var closestNeighbor *hnswVertex

		entrypoint.edgeMutexes[level].RLock()
		for neighbor, _ := range entrypoint.edges[level] {
			if neighbor.isDeleted() {
				continue
			}
			if distance := xx.distancer.Distance(query, neighbor.vector); distance < minDistance {
				minDistance = distance
				closestNeighbor = neighbor
			}
		}
		entrypoint.edgeMutexes[level].RUnlock()

		if closestNeighbor == nil {
			break
		}
		entrypoint = closestNeighbor
	}

	return entrypoint, minDistance
}

func (xx *Hnsw) searchLevel(query edge.Vector, entrypoint *hnswVertex, ef, level int) PriorityQueue {
	entrypointDistance := xx.distancer.Distance(query, entrypoint.vector)
	pqItem := NewPriorityQueueItem(entrypointDistance, entrypoint)
	candidateVertices := NewMinPriorityQueue(pqItem)
	resultVertices := NewMaxPriorityQueue(pqItem)

	visitedVertices := make(map[*hnswVertex]struct{}, ef*xx.config.mMax0)
	visitedVertices[entrypoint] = struct{}{}

	for candidateVertices.Len() > 0 {
		candidateItem := candidateVertices.Pop()
		candidate := candidateItem.Value().(*hnswVertex)
		lowerBound := resultVertices.Peek().Priority()

		if candidateItem.Priority() > lowerBound {
			break
		}

		candidate.edgeMutexes[level].RLock()
		for neighbor, _ := range candidate.edges[level] {
			if neighbor.isDeleted() {
				continue
			}
			if _, exists := visitedVertices[neighbor]; exists {
				continue
			}
			visitedVertices[neighbor] = struct{}{}

			distance := xx.distancer.Distance(query, neighbor.vector)
			if (distance < lowerBound) || (resultVertices.Len() < ef) {
				pqItem := NewPriorityQueueItem(distance, neighbor)
				candidateVertices.Push(pqItem)
				resultVertices.Push(pqItem)

				if resultVertices.Len() > ef {
					resultVertices.Pop()
				}
			}
		}
		candidate.edgeMutexes[level].RUnlock()
	}

	// MaxPriorityQueue
	return resultVertices
}

func (xx *Hnsw) selectNeighbors(neighbors PriorityQueue, k int) PriorityQueue {
	for neighbors.Len() > k {
		neighbors.Pop()
	}

	return neighbors
}

func (xx *Hnsw) selectNeighborsHeuristic(query edge.Vector, neighbors PriorityQueue, k, level int, extendCandidates, keepPruned bool) PriorityQueue {
	candidateVertices := neighbors.Reverse() // MinPriorityQueue

	existingCandidatesSize := neighbors.Len()
	if extendCandidates {
		existingCandidatesSize += neighbors.Len() * xx.config.mMax0
	}
	existingCandidates := make(map[*hnswVertex]struct{}, existingCandidatesSize)
	for _, neighbor := range neighbors.ToSlice() {
		existingCandidates[neighbor.Value().(*hnswVertex)] = struct{}{}
	}

	if extendCandidates {
		for neighbors.Len() > 0 {
			candidate := neighbors.Pop().Value().(*hnswVertex)

			candidate.edgeMutexes[level].RLock()
			for neighbor, _ := range candidate.edges[level] {
				if neighbor.isDeleted() {
					continue
				}
				if _, exists := existingCandidates[neighbor]; exists {
					continue
				}
				existingCandidates[neighbor] = struct{}{}

				distance := xx.distancer.Distance(query, neighbor.vector)
				candidateVertices.Push(NewPriorityQueueItem(distance, neighbor))
			}
			candidate.edgeMutexes[level].RUnlock()
		}
	}

	result := NewMaxPriorityQueue()
	for (candidateVertices.Len() > 0) && (result.Len() < k) {
		result.Push(candidateVertices.Pop())
	}

	if keepPruned {
		for candidateVertices.Len() > 0 {
			if result.Len() >= k {
				break
			}
			result.Push(candidateVertices.Pop())
		}
	}

	return result
}

func (xx *Hnsw) pruneNeighbors(vertex *hnswVertex, k, level int) {
	neighborsQueue := NewMaxPriorityQueue()

	vertex.edgeMutexes[level].RLock()
	for neighbor, distance := range vertex.edges[level] {
		if neighbor.isDeleted() {
			continue
		}
		neighborsQueue.Push(NewPriorityQueueItem(distance, neighbor))
	}
	vertex.edgeMutexes[level].RUnlock()

	switch xx.config.searchAlgorithm {
	case HnswSearchSimple:
		neighborsQueue = xx.selectNeighbors(neighborsQueue, k)
	case HnswSearchHeuristic:
		neighborsQueue = xx.selectNeighborsHeuristic(vertex.vector, neighborsQueue, k, level, xx.config.heuristicExtendCandidates, xx.config.heuristicKeepPruned)
	}

	newNeighbors := make(hnswEdgeSet, neighborsQueue.Len())
	for _, item := range neighborsQueue.ToSlice() {
		newNeighbors[item.Value().(*hnswVertex)] = item.Priority()
	}

	vertex.setEdges(level, newNeighbors)
}

func (xx *Hnsw) BytesSize() uint64 {
	maxLevel := 10
	if entrypoint := (*hnswVertex)(atomic.LoadPointer(&xx.entrypoint)); entrypoint != nil {
		maxLevel = entrypoint.level
	}

	mutb := float64(HNSW_VERTEX_MUTEX_BYTES)
	var pointersSize float64 = float64(xx.config.mMax0*HNSW_VERTEX_EDGE_BYTES) + mutb
	for i := 1; i < maxLevel; i++ {
		pointersSize += (float64(xx.config.mMax*HNSW_VERTEX_EDGE_BYTES) + mutb) * math.Exp(float64(i)/-float64(xx.config.levelMultiplier))
	}

	verticesDataSize := atomic.LoadUint64(&xx.bytesSize)
	return uint64(math.Floor(float64(xx.Len())*pointersSize)) + verticesDataSize
}
