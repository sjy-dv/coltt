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

import "math"

var (
	ErrCollectionNotFound = "collection: %s not found"
	panicr                = "panic %v"
	ErrCollectionExists   = "collection: %s is already exists"
	ErrCollectionNotLoad  = "collection: %s is not loaded in memory"
	ErrQuantizedFailed    = "quantized failed vector : "
	edgeData              = "./data_dir/%s-edge.cdat"
	edgeIndex             = "./data_dir/%s-edge.bin"
	edgeVector            = "./data_dir/%s-vec-edge.cdat"
	edgeConfig            = "./data_dir/%s-edge_conf.json"
	collectionEdgeJson    = "./data_dir/collection-edge.json"
	TargetIdNotFound      = "NodeID: %d is not found"
)

const (
	COSINE            = "cosine"
	EUCLIDEAN         = "euclidean"
	NONE_QAUNTIZATION = "none"
	F16_QUANTIZATION  = "f16"
	F8_QUANTIZATION   = "f8"
	BF16_QUANTIZATION = "bf16"
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
	var norm float32
	out := make([]float32, len(v))
	for i := range v {
		norm += v[i] * v[i]
	}
	if norm == 0 {
		v = out
		return
	}

	norm = float32(math.Sqrt(float64(norm)))
	for i := range v {
		out[i] = v[i] / norm
	}
	v = out
}

func (v Vector) Dimensions() int {
	return len(v)
}
