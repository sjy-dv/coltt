package hnsw

import (
	"sync"
	"sync/atomic"

	"github.com/google/uuid"
	"github.com/sjy-dv/vemoo/pkg/gomath"
)

const HNSW_VERTEX_EDGE_BYTES = 8 + 4
const HNSW_VERTEX_MUTEX_BYTES = 24

type hnswEdgeSet map[*hnswVertex]float32

type hnswVertex struct {
	id          uuid.UUID
	vector      gomath.Vector
	metadata    Metadata
	level       int
	deleted     uint32
	edges       []hnswEdgeSet
	edgeMutexes []*sync.RWMutex
}

func newHnswVertex(id uuid.UUID, vector gomath.Vector, metadata Metadata, level int) *hnswVertex {
	vertex := &hnswVertex{
		id:       id,
		vector:   vector,
		metadata: metadata,
		level:    level,
		deleted:  0,
	}
	vertex.setLevel(level)

	return vertex
}

func (this *hnswVertex) Id() uuid.UUID {
	return this.id
}

func (this *hnswVertex) Vector() gomath.Vector {
	return this.vector
}

func (this *hnswVertex) Metadata() Metadata {
	return this.metadata
}

func (this *hnswVertex) Level() int {
	return this.level
}

func (this *hnswVertex) isDeleted() bool {
	return atomic.LoadUint32(&this.deleted) == 1
}

func (this *hnswVertex) setDeleted() {
	atomic.StoreUint32(&this.deleted, 1)
}

func (this *hnswVertex) setLevel(level int) {
	this.edges = make([]hnswEdgeSet, level+1)
	this.edgeMutexes = make([]*sync.RWMutex, level+1)

	for i := 0; i <= level; i++ {
		this.edges[i] = make(hnswEdgeSet)
		this.edgeMutexes[i] = &sync.RWMutex{}
	}
}

func (this *hnswVertex) edgesCount(level int) int {
	defer this.edgeMutexes[level].RUnlock()
	this.edgeMutexes[level].RLock()

	return len(this.edges[level])
}

func (this *hnswVertex) addEdge(level int, edge *hnswVertex, distance float32) {
	defer this.edgeMutexes[level].Unlock()
	this.edgeMutexes[level].Lock()

	this.edges[level][edge] = distance
}

func (this *hnswVertex) removeEdge(level int, edge *hnswVertex) {
	defer this.edgeMutexes[level].Unlock()
	this.edgeMutexes[level].Lock()

	delete(this.edges[level], edge)
}

func (this *hnswVertex) getEdges(level int) hnswEdgeSet {
	defer this.edgeMutexes[level].RUnlock()
	this.edgeMutexes[level].RLock()

	return this.edges[level]
}

func (this *hnswVertex) setEdges(level int, edges hnswEdgeSet) {
	defer this.edgeMutexes[level].Unlock()
	this.edgeMutexes[level].Lock()

	this.edges[level] = edges
}

func (this *hnswVertex) bytesSize() uint64 {
	return 16 + gomath.VECTOR_COMPONENT_BYTES_SIZE*uint64(len(this.vector)) + this.metadata.bytesSize()
}
