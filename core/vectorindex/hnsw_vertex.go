package vectorindex

import (
	"sync"
	"sync/atomic"

	"github.com/sjy-dv/nnv/edge"
)

const HNSW_VERTEX_EDGE_BYTES = 8 + 4
const HNSW_VERTEX_MUTEX_BYTES = 24

type hnswEdgeSet map[*hnswVertex]float32

type hnswVertex struct {
	id          uint64
	vector      edge.Vector
	level       int
	metadata    Metadata
	deleted     uint32
	edges       []hnswEdgeSet
	edgeMutexes []*sync.RWMutex
}

func newHnswVertex(id uint64, vector edge.Vector, metadata Metadata, level int) *hnswVertex {
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

func (xx *hnswVertex) Id() uint64 {
	return xx.id
}

func (xx *hnswVertex) Vector() edge.Vector {
	return xx.vector
}

func (xx *hnswVertex) Metadata() Metadata {
	return xx.metadata
}

func (xx *hnswVertex) Level() int {
	return xx.level
}

func (xx *hnswVertex) isDeleted() bool {
	return atomic.LoadUint32(&xx.deleted) == 1
}

func (xx *hnswVertex) setDeleted() {
	atomic.StoreUint32(&xx.deleted, 1)
}

func (xx *hnswVertex) setLevel(level int) {
	xx.edges = make([]hnswEdgeSet, level+1)
	xx.edgeMutexes = make([]*sync.RWMutex, level+1)

	for i := 0; i <= level; i++ {
		xx.edges[i] = make(hnswEdgeSet)
		xx.edgeMutexes[i] = &sync.RWMutex{}
	}
}

func (xx *hnswVertex) edgesCount(level int) int {
	defer xx.edgeMutexes[level].RUnlock()
	xx.edgeMutexes[level].RLock()

	return len(xx.edges[level])
}

func (xx *hnswVertex) addEdge(level int, edge *hnswVertex, distance float32) {
	defer xx.edgeMutexes[level].Unlock()
	xx.edgeMutexes[level].Lock()

	xx.edges[level][edge] = distance
}

func (xx *hnswVertex) removeEdge(level int, edge *hnswVertex) {
	defer xx.edgeMutexes[level].Unlock()
	xx.edgeMutexes[level].Lock()

	delete(xx.edges[level], edge)
}

func (xx *hnswVertex) getEdges(level int) hnswEdgeSet {
	defer xx.edgeMutexes[level].RUnlock()
	xx.edgeMutexes[level].RLock()

	return xx.edges[level]
}

func (xx *hnswVertex) setEdges(level int, edges hnswEdgeSet) {
	defer xx.edgeMutexes[level].Unlock()
	xx.edgeMutexes[level].Lock()

	xx.edges[level] = edges
}

func (xx *hnswVertex) bytesSize() uint64 {
	//  uint64 = 8byte
	// float32 => 4 byte x vector len
	return 8 + 4*uint64(len(xx.vector)) + xx.metadata.byteSize()
}
