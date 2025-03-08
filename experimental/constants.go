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

package experimental

import "math"

type Vector []float32

func (v Vector) Dimensions() int {
	return len(v)
}

type VertexEdge struct {
	MultiVectors map[string]Vector
	Metadata     map[string]any
}

var (
	panicr = "panic %v"
)

const (
	COSINE                 = "cosine"
	EUCLIDEAN              = "euclidean"
	NONE_QAUNTIZATION      = "none"
	F16_QUANTIZATION       = "f16"
	F8_QUANTIZATION        = "f8"
	BF16_QUANTIZATION      = "bf16"
	T_COSINE               = "cosine-dot"
	VERTEX_SHARD_COUNT int = 16
)

func Normalize(v []float32) []float32 {
	var norm float32
	out := make([]float32, len(v))
	for i := range v {
		norm += v[i] * v[i]
	}
	if norm == 0 {
		return out
	}

	norm = float32(math.Sqrt(float64(norm)))
	for i := range v {
		out[i] = v[i] / norm
	}

	return out
}
