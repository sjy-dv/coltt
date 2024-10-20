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

package sse

import (
	"math"
	"unsafe"
)

//go:noescape
func _euclidean_distance_squared(len, a, b, result unsafe.Pointer)

func EuclideanDistance(a, b []float32) float32 {
	var result float32
	_euclidean_distance_squared(unsafe.Pointer(uintptr(len(a))), unsafe.Pointer(&a[0]), unsafe.Pointer(&b[0]), unsafe.Pointer(&result))
	return float32(math.Sqrt(float64(result)))
}

//go:noescape
func _manhattan_distance(len, a, b, result unsafe.Pointer)

func ManhattanDistance(a, b []float32) float32 {
	var result float32
	_manhattan_distance(unsafe.Pointer(uintptr(len(a))), unsafe.Pointer(&a[0]), unsafe.Pointer(&b[0]), unsafe.Pointer(&result))
	return result
}

//go:noescape
func _cosine_similarity_dot_norm(len, a, b, dot, norm_squared unsafe.Pointer)

func CosineDistance(a, b []float32) float32 {
	var dot float32
	var norm_squared float32
	_cosine_similarity_dot_norm(unsafe.Pointer(uintptr(len(a))), unsafe.Pointer(&a[0]), unsafe.Pointer(&b[0]), unsafe.Pointer(&dot), unsafe.Pointer(&norm_squared))

	return 1.0 - dot/float32(math.Sqrt(float64(norm_squared)))
}
