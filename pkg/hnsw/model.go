package hnsw

import (
	"sync"

	"github.com/sjy-dv/nnv/pkg/distance"
	"github.com/sjy-dv/nnv/pkg/gomath"
)

// metadata must contain _id <- is find key
// or user add id but, user forced nodeid
type Node struct {
	LinkNodes [][]uint32
	Vectors   gomath.Vector
	Layer     int // hnsw layer tree
	Id        uint32
	Timestamp uint64 // check node put order
	Metadata  map[string]interface{}
	IsEmpty   bool
}

type NodeList struct {
	Nodes []Node
	rmu   sync.RWMutex
}

type HnswConfig struct {
	Efconstruction int
	M              int
	Mmax           int
	Mmax0          int
	Ml             float64
	Ep             int64
	MaxLevel       int
	Dim            uint32
	DistanceType   string
	Heuristic      bool
	BucketName     string // using seperate vector or find prefix kv
	EmptyNodes     []uint32
}

type Hnsw struct {
	Efconstruction int
	M              int
	Mmax           int
	Mmax0          int
	Ml             float64
	Ep             int64
	MaxLevel       int
	Dim            uint32
	Heuristic      bool
	Space          distance.Space
	DistanceType   string
	NodeList       NodeList
	BucketName     string   // using seperate vector or find prefix kv
	EmptyNodes     []uint32 // restore empty node link
	rmu            sync.RWMutex
	Wg             sync.WaitGroup
}

// type IndexFilter map[string]IndexType

// type IndexType struct {
//     Type string
//     String *
// }

type HnswBucket struct {
	Buckets     map[string]*Hnsw // bucket managing multi-hnsw nodes
	BucketGroup map[string]bool
	rmu         sync.RWMutex
}

type SearchQuery struct {
	Id int
	Qp []float32
}

type SearchResults struct {
	Id             int
	BestCandidates PriorityQueue
}
