package space

import "github.com/sjy-dv/vemoo/pkg/gomath"

type nativeSpaceImpl struct{}

func (nativeSpaceImpl) EuclideanDistance(a, b gomath.Vector) float32 {
	var distance float32
	for i := 0; i < len(a); i++ {
		distance += gomath.Square(a[i] - b[i])
	}

	return gomath.Sqrt(distance)
}

func (nativeSpaceImpl) ManhattanDistance(a, b gomath.Vector) float32 {
	var distance float32
	for i := 0; i < len(a); i++ {
		distance += gomath.Abs(a[i] - b[i])
	}

	return distance
}

func (nativeSpaceImpl) CosineDistance(a, b gomath.Vector) float32 {
	var dot float32
	var aNorm float32
	var bNorm float32
	for i := 0; i < len(a); i++ {
		dot += a[i] * b[i]
		aNorm += gomath.Square(a[i])
		bNorm += gomath.Square(b[i])
	}

	return 1.0 - dot/(gomath.Sqrt(aNorm)*gomath.Sqrt(bNorm))
}
