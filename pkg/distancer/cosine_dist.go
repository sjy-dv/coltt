package distancer

import (
	"github.com/pkg/errors"
)

type CosineDistance struct {
	a []float32
}

func (d *CosineDistance) Distance(b []float32) (float32, error) {
	if len(d.a) != len(b) {
		return 0, errors.Wrapf(ErrVectorLength, "%d vs %d",
			len(d.a), len(b))
	}

	dist := 1 - dotProductImplementation(d.a, b)
	return dist, nil
}

type CosineDistanceProvider struct{}

func NewCosineDistanceProvider() CosineDistanceProvider {
	return CosineDistanceProvider{}
}

func (d CosineDistanceProvider) SingleDist(a, b []float32) (float32, error) {
	if len(a) != len(b) {
		return 0, errors.Wrapf(ErrVectorLength, "%d vs %d",
			len(a), len(b))
	}

	prod := 1 - dotProductImplementation(a, b)

	return prod, nil
}

func (d CosineDistanceProvider) Type() string {
	return "cosine-dot"
}

func (d CosineDistanceProvider) New(a []float32) Distancer {
	return &CosineDistance{a: a}
}

func (d CosineDistanceProvider) Step(x, y []float32) float32 {
	var sum float32
	for i := range x {
		sum += x[i] * y[i]
	}

	return sum
}

func (d CosineDistanceProvider) Wrap(x float32) float32 {
	return 1 - x
}
