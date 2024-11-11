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

// edge_quantization.go
package edge

import (
	"encoding/binary"
	"math"

	"github.com/viterin/vek/vek32"
)

type Quantization[L any] interface {
	Similarity(x, y L) float32
	Lower(v Vector) (L, error)
	Marshal(to []byte, lower L) error
	Unmarshal(data []byte) (L, error)
	Name() string
	LowerSize(dim int) int
}

var _ Quantization[Vector] = NoQuantization{}

type NoQuantization struct{}

func (q NoQuantization) Similarity(x, y Vector) float32 {
	return vek32.CosineSimilarity(x, y)
}

func (q NoQuantization) Lower(v Vector) (Vector, error) {
	return v, nil
}

func (q NoQuantization) Marshal(to []byte, lower Vector) error {
	for i, n := range lower {
		u := math.Float32bits(n)
		binary.LittleEndian.PutUint32(to[i*4:], u)
	}
	return nil
}

func (q NoQuantization) Unmarshal(data []byte) (Vector, error) {
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
