package hnswpq

import (
	"sync"

	"github.com/sjy-dv/nnv/pkg/conversion"
	"github.com/sjy-dv/nnv/pkg/distancepq"
	"github.com/sjy-dv/nnv/pkg/gomath"
)

type Node struct {
	LinkNodes [][]uint64
	Vectors   gomath.Vector
	Layer     int
	Id        uint64
	IsEmpty   bool
	Centroids []uint8
}

type NodeList struct {
	Nodes []Node
	lock  sync.RWMutex
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
	DistFn         distancepq.FloatDistFunc
	DistFnName     string
	NodeList       NodeList
	CollectionName string
	EmptyNodes     []uint64
	hlock          sync.RWMutex
	Wg             sync.WaitGroup
	PQ             *productQuantizer
}

type ProductQuantizerParameters struct {
	// Number of centroids to quantize to, this is the k* parameter in the paper
	// and is often set to 255 giving 256 centroids (including 0). We are
	// limiting this to maximum of 256 (uint8) to keep the overhead of this
	// process tractable.
	NumCentroids int `json:"numCentroids" binding:"required,min=2,max=256"`
	// Number of subvectors / segments / subquantizers to use, this is the m
	// parameter in the paper and is often set to 8.
	NumSubVectors int `json:"numSubVectors" binding:"required,min=2"`
	// Number of points to use to train the quantizer, it will automatically trigger training
	// when this number of points is reached.
	TriggerThreshold int `json:"triggerThreshold" binding:"required,min=1000,max=10000"`
}

type productQuantizedPoint struct {
	id          uint64
	Vector      []float32
	CentroidIds []uint8
	isDirty     bool
}

func (p *productQuantizedPoint) Id() uint64 {
	return p.id
}

func (p *productQuantizedPoint) IdFromKey(key []byte) (uint64, bool) {
	return conversion.NodeIdFromKey(key, 'v')
}

func (p *productQuantizedPoint) SizeInMemory() int64 {
	return int64(8 + 4*len(p.Vector) + len(p.CentroidIds))
}

func (p *productQuantizedPoint) CheckAndClearDirty() bool {
	dirty := p.isDirty
	p.isDirty = false
	return dirty
}

type PointIdDistFn func(y *productQuantizedPoint) float32
