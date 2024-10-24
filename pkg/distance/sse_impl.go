package distance

import (
	"github.com/sjy-dv/nnv/pkg/distance/simd/sse"
	"github.com/sjy-dv/nnv/pkg/gomath"
)

type sseSpaceImpl struct{}

func (sseSpaceImpl) EuclideanDistance(a, b gomath.Vector) float32 {
	return sse.EuclideanDistance(a, b)
}

func (sseSpaceImpl) ManhattanDistance(a, b gomath.Vector) float32 {
	return sse.ManhattanDistance(a, b)
}

func (sseSpaceImpl) CosineDistance(a, b gomath.Vector) float32 {
	return sse.CosineDistance(a, b)
}
