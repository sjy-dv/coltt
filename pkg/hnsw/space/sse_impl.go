package space

import (
	"github.com/sjy-dv/vemoo/pkg/gomath"
	"github.com/sjy-dv/vemoo/pkg/hnsw/simd/sse"
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
