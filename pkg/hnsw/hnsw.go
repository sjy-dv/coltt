package hnsw

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/google/uuid"
	"github.com/sjy-dv/vemoo/pkg/gomath"
	"github.com/sjy-dv/vemoo/pkg/hnsw/space"
	"github.com/sjy-dv/vemoo/pkg/priorityqueue"
	"github.com/sjy-dv/vemoo/pkg/sharding"
)

const VERTICES_MAP_SHARD_COUNT int = 16

var (
	ItemNotFoundError      error = errors.New("Item not found")
	ItemAlreadyExistsError error = errors.New("Item already exists")
)

type Hnsw struct {
	size      uint
	bytesSize uint64
	space     space.Space
	config    *hnswConfig

	len        uint64
	vertices   [VERTICES_MAP_SHARD_COUNT]map[uuid.UUID]*hnswVertex
	verticesMu [VERTICES_MAP_SHARD_COUNT]*sync.RWMutex

	entrypoint unsafe.Pointer
}

func NewHnsw(size uint, space space.Space, options ...HnswOption) *Hnsw {
	index := &Hnsw{
		size:   size,
		space:  space,
		config: newHnswConfig(options),

		len:        0,
		entrypoint: nil,
	}

	for i := 0; i < VERTICES_MAP_SHARD_COUNT; i++ {
		index.vertices[i] = make(map[uuid.UUID]*hnswVertex)
		index.verticesMu[i] = &sync.RWMutex{}
	}

	return index
}

func (this *Hnsw) String() string {
	return fmt.Sprintf("HNSW(dim: %d, space: %s, config={%s})", this.size, this.space, this.config)
}

func (this *Hnsw) Len() int {
	return int(atomic.LoadUint64(&this.len))
}

func (this *Hnsw) BytesSize() uint64 {
	maxLevel := 10
	if entrypoint := (*hnswVertex)(atomic.LoadPointer(&this.entrypoint)); entrypoint != nil {
		maxLevel = entrypoint.level
	}

	mutb := float64(HNSW_VERTEX_MUTEX_BYTES)
	var pointersSize float64 = float64(this.config.mMax0*HNSW_VERTEX_EDGE_BYTES) + mutb
	for i := 1; i < maxLevel; i++ {
		pointersSize += (float64(this.config.mMax*HNSW_VERTEX_EDGE_BYTES) + mutb) * math.Exp(float64(i)/-float64(this.config.levelMultiplier))
	}

	verticesDataSize := atomic.LoadUint64(&this.bytesSize)
	return uint64(math.Floor(float64(this.Len())*pointersSize)) + verticesDataSize
}

func (this *Hnsw) Insert(id uuid.UUID, value gomath.Vector, metadata Metadata, vertexLevel int) error {
	var vertex *hnswVertex
	if (*hnswVertex)(atomic.LoadPointer(&this.entrypoint)) == nil {
		vertex = newHnswVertex(id, value, metadata, 0)
		if err := this.storeVertex(vertex); err != nil {
			return err
		}
		if atomic.CompareAndSwapPointer(&this.entrypoint, nil, unsafe.Pointer(vertex)) {
			return nil
		} else {
			vertex.setLevel(vertexLevel)
		}
	} else {
		vertex = newHnswVertex(id, value, metadata, vertexLevel)
		if err := this.storeVertex(vertex); err != nil {
			return err
		}
	}

	entrypoint := (*hnswVertex)(atomic.LoadPointer(&this.entrypoint))
	minDistance := this.space.Distance(vertex.vector, entrypoint.vector)
	for l := entrypoint.level; l > vertex.level; l-- {
		entrypoint, minDistance = this.greedyClosestNeighbor(vertex.vector, entrypoint, minDistance, l)
	}

	for l := gomath.MinInt(entrypoint.level, vertex.level); l >= 0; l-- {
		neighbors := this.searchLevel(vertex.vector, entrypoint, this.config.efConstruction, l)

		switch this.config.searchAlgorithm {
		case HnswSearchSimple:
			neighbors = this.selectNeighbors(neighbors, this.config.m)
		case HnswSearchHeuristic:
			neighbors = this.selectNeighborsHeuristic(vertex.vector, neighbors, this.config.m, l, this.config.heuristicExtendCandidates, this.config.heuristicKeepPruned)
		}

		mMax := this.config.mMax
		if l == 0 {
			mMax = this.config.mMax0
		}

		for neighbors.Len() > 0 {
			item := neighbors.Pop()
			neighbor := item.Value().(*hnswVertex)
			entrypoint = neighbor

			vertex.addEdge(l, neighbor, item.Priority())
			neighbor.addEdge(l, vertex, item.Priority())

			if neighbor.edgesCount(l) > mMax {
				this.pruneNeighbors(neighbor, mMax, l)
			}
		}
	}

	entrypoint = (*hnswVertex)(atomic.LoadPointer(&this.entrypoint))
	if entrypoint != nil && vertex.level > entrypoint.level {
		atomic.CompareAndSwapPointer(&this.entrypoint, this.entrypoint, unsafe.Pointer(vertex))
	}

	return nil
}

func (this *Hnsw) Get(id uuid.UUID) (gomath.Vector, error) {
	m, mu := this.getVerticesShard(id)
	mu.RLock()
	defer mu.RUnlock()

	if vertex, exists := m[id]; exists {
		return vertex.vector, nil
	}
	return nil, ItemNotFoundError
}

func (this *Hnsw) GetVertex(id uuid.UUID) (*hnswVertex, error) {
	m, mu := this.getVerticesShard(id)
	mu.RLock()
	defer mu.RUnlock()

	if vertex, exists := m[id]; exists {
		return vertex, nil
	}
	return nil, ItemNotFoundError
}

func (this *Hnsw) Remove(id uuid.UUID) error {
	vertex, err := this.removeVertex(id)
	if err != nil {
		return err
	}

	currEntrypoint := atomic.LoadPointer(&this.entrypoint)
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
		atomic.CompareAndSwapPointer(&this.entrypoint, currEntrypoint, unsafe.Pointer(closestNeighbor))
	}

	for l := vertex.level; l >= 0; l-- {
		mMax := this.config.mMax
		if l == 0 {
			mMax = this.config.mMax0
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
			this.pruneNeighbors(neighbor, mMax, l)
		}
	}

	return nil
}

func (this *Hnsw) Search(ctx context.Context, query gomath.Vector, k uint) (SearchResult, error) {
	entrypoint := (*hnswVertex)(atomic.LoadPointer(&this.entrypoint))
	if entrypoint == nil {
		return make(SearchResult, 0), nil
	}

	minDistance := this.space.Distance(query, entrypoint.vector)
	for l := entrypoint.level; l > 0; l-- {
		entrypoint, minDistance = this.greedyClosestNeighbor(query, entrypoint, minDistance, l)
	}

	ef := gomath.MaxInt(this.config.ef, int(k))
	neighbors := this.searchLevel(query, entrypoint, ef, 0)

	switch this.config.searchAlgorithm {
	case HnswSearchSimple:
		neighbors = this.selectNeighbors(neighbors, int(k))
	case HnswSearchHeuristic:
		neighbors = this.selectNeighborsHeuristic(query, neighbors, int(k), 0, this.config.heuristicExtendCandidates, this.config.heuristicKeepPruned)
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

func (this *Hnsw) RandomLevel() int {
	return gomath.Floor(gomath.RandomExponential(this.config.levelMultiplier))
}

func (this *Hnsw) getVerticesShard(id uuid.UUID) (map[uuid.UUID]*hnswVertex, *sync.RWMutex) {
	shardIdx := sharding.ShardTraffic(id, uint64(VERTICES_MAP_SHARD_COUNT))
	return this.vertices[shardIdx], this.verticesMu[shardIdx]
}

func (this *Hnsw) storeVertex(vertex *hnswVertex) error {
	m, mu := this.getVerticesShard(vertex.id)
	defer mu.Unlock()
	mu.Lock()

	if _, exists := m[vertex.id]; exists {
		return ItemAlreadyExistsError
	}

	m[vertex.id] = vertex
	atomic.AddUint64(&this.len, 1)
	atomic.AddUint64(&this.bytesSize, vertex.bytesSize())
	return nil
}

func (this *Hnsw) removeVertex(id uuid.UUID) (*hnswVertex, error) {
	m, mu := this.getVerticesShard(id)
	defer mu.Unlock()
	mu.Lock()

	if vertex, exists := m[id]; exists {
		delete(m, id)
		atomic.AddUint64(&this.len, ^uint64(0))
		atomic.AddUint64(&this.bytesSize, ^uint64(vertex.bytesSize()-1))
		vertex.setDeleted()
		return vertex, nil
	}

	return nil, ItemNotFoundError
}

func (this *Hnsw) greedyClosestNeighbor(query gomath.Vector, entrypoint *hnswVertex, minDistance float32, level int) (*hnswVertex, float32) {
	for {
		var closestNeighbor *hnswVertex

		entrypoint.edgeMutexes[level].RLock()
		for neighbor, _ := range entrypoint.edges[level] {
			if neighbor.isDeleted() {
				continue
			}
			if distance := this.space.Distance(query, neighbor.vector); distance < minDistance {
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

func (this *Hnsw) searchLevel(query gomath.Vector, entrypoint *hnswVertex, ef, level int) priorityqueue.PriorityQueue {
	entrypointDistance := this.space.Distance(query, entrypoint.vector)

	pqItem := priorityqueue.NewPriorityQueueItem(entrypointDistance, entrypoint)
	candidateVertices := priorityqueue.NewMinPriorityQueue(pqItem)
	resultVertices := priorityqueue.NewMaxPriorityQueue(pqItem)

	visitedVertices := make(map[*hnswVertex]struct{}, ef*this.config.mMax0)
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

			distance := this.space.Distance(query, neighbor.vector)
			if (distance < lowerBound) || (resultVertices.Len() < ef) {
				pqItem := priorityqueue.NewPriorityQueueItem(distance, neighbor)
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

func (this *Hnsw) selectNeighbors(neighbors priorityqueue.PriorityQueue, k int) priorityqueue.PriorityQueue {
	for neighbors.Len() > k {
		neighbors.Pop()
	}

	return neighbors
}

func (this *Hnsw) selectNeighborsHeuristic(query gomath.Vector, neighbors priorityqueue.PriorityQueue, k, level int, extendCandidates, keepPruned bool) priorityqueue.PriorityQueue {
	candidateVertices := neighbors.Reverse() // MinPriorityQueue

	existingCandidatesSize := neighbors.Len()
	if extendCandidates {
		existingCandidatesSize += neighbors.Len() * this.config.mMax0
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

				distance := this.space.Distance(query, neighbor.vector)
				candidateVertices.Push(priorityqueue.NewPriorityQueueItem(distance, neighbor))
			}
			candidate.edgeMutexes[level].RUnlock()
		}
	}

	result := priorityqueue.NewMaxPriorityQueue()
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

func (this *Hnsw) pruneNeighbors(vertex *hnswVertex, k, level int) {
	neighborsQueue := priorityqueue.NewMaxPriorityQueue()

	vertex.edgeMutexes[level].RLock()
	for neighbor, distance := range vertex.edges[level] {
		if neighbor.isDeleted() {
			continue
		}
		neighborsQueue.Push(priorityqueue.NewPriorityQueueItem(distance, neighbor))
	}
	vertex.edgeMutexes[level].RUnlock()

	switch this.config.searchAlgorithm {
	case HnswSearchSimple:
		neighborsQueue = this.selectNeighbors(neighborsQueue, k)
	case HnswSearchHeuristic:
		neighborsQueue = this.selectNeighborsHeuristic(vertex.vector, neighborsQueue, k, level, this.config.heuristicExtendCandidates, this.config.heuristicKeepPruned)
	}

	newNeighbors := make(hnswEdgeSet, neighborsQueue.Len())
	for _, item := range neighborsQueue.ToSlice() {
		newNeighbors[item.Value().(*hnswVertex)] = item.Priority()
	}

	vertex.setEdges(level, newNeighbors)
}
