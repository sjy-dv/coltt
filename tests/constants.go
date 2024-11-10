// constants.go
package edge

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"

	"github.com/viterin/vek/vek32"
)

var (
	ErrAlreadyBuilt = errors.New("Already built the index")
	ErrIDNotFound   = errors.New("ID not found")
)

type ID uint64

type Basis []Vector

type Vector []float32

func (v Vector) Clone() Vector {
	out := make([]float32, len(v))
	copy(out, v)
	return out
}

func (v Vector) Normalize() {
	factor := vek32.Norm(v)
	vek32.DivNumber_Inplace(v, factor)
}

func (v Vector) Dimensions() int {
	return len(v)
}

func (v Vector) CosineSimilarity(other Vector) float32 {
	return vek32.CosineSimilarity(v, other)
}

func NewRandVector(dim int, rng *rand.Rand) Vector {
	out := make([]float32, dim)
	for i := 0; i < dim; i++ {
		if rng != nil {
			out[i] = float32(rng.NormFloat64())
		} else {
			out[i] = float32(rand.NormFloat64())
		}
	}
	factor := vek32.Norm(out)
	vek32.DivNumber_Inplace(out, factor)
	return out
}

func NewRandVectorSet(n int, dim int, rng *rand.Rand) []Vector {
	out := make([]Vector, n)
	for i := 0; i < n; i++ {
		out[i] = NewRandVector(dim, rng)
		out[i].Normalize()
	}
	return out
}

// Modified Gram-Schmidt (Same as before)
func orthonormalize(basis Basis) {
	buf := make([]float32, len(basis[0]))
	cur := basis[0]
	for i := 1; i < len(basis); i++ {
		for j := i; j < len(basis); j++ {
			dot := vek32.Dot(basis[j], cur)
			vek32.MulNumber_Into(buf, cur, dot)
			vek32.Sub_Inplace(basis[j], buf)
			basis[j].Normalize()
		}
		cur = basis[i]
	}
}

func debugPrintBasis(basis Basis) {
	for i := 0; i < len(basis); i++ {
		sim := make([]any, len(basis))
		for j := 0; j < len(basis); j++ {
			sim[j] = vek32.CosineSimilarity(basis[i], basis[j])
		}
		pattern := strings.Repeat("%+.15f  ", len(basis))
		fmt.Printf(pattern+"\n", sim...)
	}
}
