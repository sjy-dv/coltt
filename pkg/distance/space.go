package distance

import (
	"github.com/klauspost/cpuid"
	"github.com/sjy-dv/nnv/pkg/gomath"
)

type SpaceImpl interface {
	EuclideanDistance(gomath.Vector, gomath.Vector) float32
	ManhattanDistance(gomath.Vector, gomath.Vector) float32
	CosineDistance(gomath.Vector, gomath.Vector) float32
}

type Space interface {
	Distance(gomath.Vector, gomath.Vector) float32
}

type space struct {
	impl SpaceImpl
}

func newSpace() space {
	if cpuid.CPU.AVX() {
		return space{impl: avxSpaceImpl{}}
	}
	if cpuid.CPU.SSE() {
		return space{impl: sseSpaceImpl{}}
	}

	return space{impl: nativeSpaceImpl{}}
}

type Euclidean struct{ space }

type Manhattan struct{ space }

type Cosine struct{ space }

func NewEuclidean() Space {
	return &Euclidean{newSpace()}
}

func (this *Euclidean) Distance(a, b gomath.Vector) float32 {
	return this.impl.EuclideanDistance(a, b)
}

func (this *Euclidean) String() string {
	return "euclidean"
}

func NewManhattan() Space {
	return &Manhattan{newSpace()}
}

func (this *Manhattan) Distance(a, b gomath.Vector) float32 {
	return this.impl.ManhattanDistance(a, b)
}

func (this *Manhattan) String() string {
	return "manhattan"
}

func NewCosine() Space {
	return &Cosine{newSpace()}
}

func (this *Cosine) Distance(a, b gomath.Vector) float32 {
	return gomath.Abs(this.impl.CosineDistance(a, b))
}

func (this *Cosine) String() string {
	return "cosine"
}
