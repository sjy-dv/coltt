package space

import (
	"github.com/sjy-dv/vemoo/pkg/gomath"
	"github.com/sjy-dv/vemoo/pkg/hnsw/simd/avx"
)

type avxSpaceImpl struct{}

func (avxSpaceImpl) EuclideanDistance(a, b gomath.Vector) float32 {
	return avx.EuclideanDistance(a, b)
}

func (avxSpaceImpl) ManhattanDistance(a, b gomath.Vector) float32 {
	return avx.ManhattanDistance(a, b)
}

func (avxSpaceImpl) CosineDistance(a, b gomath.Vector) float32 {
	return avx.CosineDistance(a, b)
}
