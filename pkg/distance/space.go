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

package distance

import (
	"github.com/klauspost/cpuid"
	"github.com/sjy-dv/coltt/pkg/gomath"
)

type SpaceImpl interface {
	EuclideanDistance([]float32, []float32) float32
	ManhattanDistance([]float32, []float32) float32
	CosineDistance([]float32, []float32) float32
}

type Space interface {
	Distance([]float32, []float32) float32
	Type() string
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

func (this *Euclidean) Distance(a, b []float32) float32 {
	return this.impl.EuclideanDistance(a, b)
}

func (this *Euclidean) String() string {
	return "euclidean"
}

func (this *Euclidean) Type() string {
	return "l2-squared"
}

func NewManhattan() Space {
	return &Manhattan{newSpace()}
}

func (this *Manhattan) Distance(a, b []float32) float32 {
	return this.impl.ManhattanDistance(a, b)
}

func (this *Manhattan) String() string {
	return "manhattan"
}

func (this *Manhattan) Type() string {
	return "manhattan"
}

func NewCosine() Space {
	return &Cosine{newSpace()}
}

func (this *Cosine) Distance(a, b []float32) float32 {
	return gomath.Abs(this.impl.CosineDistance(a, b))
}

func (this *Cosine) String() string {
	return "cosine"
}

func (this *Cosine) Type() string {
	return "cosine-dot"
}
