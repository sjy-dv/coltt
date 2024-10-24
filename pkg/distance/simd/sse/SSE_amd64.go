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
