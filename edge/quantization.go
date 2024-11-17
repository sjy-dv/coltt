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
	"encoding/binary"
	"math"

	"github.com/sjy-dv/nnv/pkg/distance"
	"github.com/sjy-dv/nnv/pkg/gomath"
)

type Quantization[T any] interface {
	Similarity(x, y T, dist distance.Space) float32
	Lower(v gomath.Vector) (T, error)
	Name() string
	LowerSize(dim int) int
}

type QuantizationType interface {
	gomath.Vector | float16Vec |
		bfloat16Vec | float8Vec
}

var _ Quantization[gomath.Vector] = NoQuantization{}

type NoQuantization struct{}

func (q NoQuantization) Similarity(x, y gomath.Vector, dist distance.Space) float32 {
	return dist.Distance(x, y)
}

func (q NoQuantization) Lower(v gomath.Vector) (gomath.Vector, error) {
	return v, nil
}

func (q NoQuantization) Marshal(to []byte, lower gomath.Vector) error {
	for i, n := range lower {
		u := math.Float32bits(n)
		binary.LittleEndian.PutUint32(to[i*4:], u)
	}
	return nil
}

func (q NoQuantization) Unmarshal(data []byte) (gomath.Vector, error) {
	out := make([]float32, len(data)>>2)
	for i := 0; i < len(data); i += 4 {
		bits := binary.LittleEndian.Uint32(data[i:])
		out[i>>2] = math.Float32frombits(bits)
	}
	return out, nil
}

func (q NoQuantization) Name() string {
	return "none"
}

func (q NoQuantization) LowerSize(dim int) int {
	return 4 * dim
}
