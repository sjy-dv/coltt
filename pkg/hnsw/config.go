package hnsw

import (
	"math"

	"github.com/sjy-dv/nnv/pkg/distance"
)

func DefaultConfig(dim uint32, bucketName string) *HnswConfig {
	return &HnswConfig{
		Efconstruction: 200,
		M:              16,
		Mmax:           16,
		Mmax0:          32,
		Ml:             1 / math.Log(1.0*float64(16)),
		Ep:             0,
		MaxLevel:       0,
		Heuristic:      true,
		Space:          distance.NewEuclidean(),
		Dim:            dim,
		BucketName:     bucketName,
	}
}
