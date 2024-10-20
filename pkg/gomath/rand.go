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

package gomath

import "math/rand"

func RandomDistinctInts(n, max int) []int {
	result := make([]int, n)
	result[0] = rand.Intn(max)

	i := 1
	for i < n {
		candidate := rand.Intn(max)
		if candidate != result[i-1] {
			result[i] = candidate
			i++
		}
	}

	return result
}

func RandomUniform() float32 {
	return float32(rand.Float32())
}

func RandomExponential(lambda float32) float32 {
	return -Log(rand.Float32()) * lambda
}

func RandomUniformVector(size int) Vector {
	vec := make(Vector, size)
	for i := 0; i < size; i++ {
		vec[i] = RandomUniform()
	}
	return vec
}

func RandomStandardNormalVector(size int) Vector {
	vec := make(Vector, size)
	for i := 0; i < size; i++ {
		vec[i] = float32(rand.NormFloat64())
	}
	return vec
}

func RandomNormalVector(size int, mu, sigma float32) Vector {
	vec := make(Vector, size)
	for i := 0; i < size; i++ {
		vec[i] = float32(rand.NormFloat64())*sigma + mu
	}
	return vec
}
