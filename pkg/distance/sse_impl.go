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
	"github.com/sjy-dv/nnv/pkg/distance/simd/sse"
	"github.com/sjy-dv/nnv/pkg/gomath"
)

type sseSpaceImpl struct{}

func (sseSpaceImpl) EuclideanDistance(a, b gomath.Vector) float32 {
	return sse.EuclideanDistance(a, b)
}

func (sseSpaceImpl) ManhattanDistance(a, b gomath.Vector) float32 {
	return sse.ManhattanDistance(a, b)
}

func (sseSpaceImpl) CosineDistance(a, b gomath.Vector) float32 {
	return sse.CosineDistance(a, b)
}
