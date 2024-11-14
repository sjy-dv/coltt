package highmem

import (
	"github.com/sjy-dv/nnv/pkg/gomath"
	"github.com/sjy-dv/nnv/pkg/hnsw"
	"github.com/sjy-dv/nnv/pkg/hnswpq"
)

type HNSWImpl interface {
	CreateCollection(
		collectionName string,
		config hnsw.HnswConfig,
		params hnswpq.ProductQuantizerParameters) error
	Genesis(collectionName string, config hnsw.HnswConfig)
	DropCollection(collectionName string) error

	//vector
	Insert(collectionName string, commitID uint64, vec gomath.Vector) error
	// Search(collectionName string, vec []float32, topCandidates)
}
