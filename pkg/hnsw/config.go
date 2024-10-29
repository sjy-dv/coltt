package hnsw

import (
	"math"
)

func DefaultConfig(dim uint32, bucketName string, dist string) HnswConfig {
	return HnswConfig{
		Efconstruction: 200,
		M:              16,
		Mmax:           16,
		Mmax0:          32,
		Ml:             1 / math.Log(1.0*float64(16)),
		Ep:             0,
		MaxLevel:       0,
		Heuristic:      true,
		DistanceType:   dist,
		Dim:            dim,
		BucketName:     bucketName,
	}
}
