package distance

import (
	"github.com/sjy-dv/nnv/pkg/distance/simd/avx"
	"github.com/sjy-dv/nnv/pkg/gomath"
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
