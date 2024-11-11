// Licensed to sjy-dv under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. sjy-dv licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package edge

import (
	"fmt"
	"math/rand/v2"
	"strings"

	"github.com/viterin/vek/vek32"
)

var (
	ErrCollectionNotFound = "collection: %s not found"
	panicr                = "panic %v"
	ErrCollectionExists   = "collection: %s is already exists"
	ErrQuantizedFailed    = "quantized failed vector : "
	edgeData              = "./data_dir/%s-edge.cdat"
	edgeIndex             = "./data_dir/%s-edge.bin"
	edgeVector            = "./data_dir/%s-edge.cdat"
	edgeConfig            = "./data_dir/%s-edge_conf.json"
	collectionEdgeJson    = "./data_dir/collection-edge.json"
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
