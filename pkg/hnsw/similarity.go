package hnsw

import "math"

// ------------------------------
// Similarity Functions
// ------------------------------

// DotProduct computes the dot product of two vectors.
func DotProduct(a, b []float32) float32 {
	if len(a) != len(b) {
		panic("vectors must be of the same length")
	}
	var dP float32
	for i := 0; i < len(a); i++ {
		dP += a[i] * b[i]
	}
	return dP
}

// CosineSimilarity computes the cosine similarity between two vectors.
func CosineSimilarity(a, b []float32) float32 {
	dp := DotProduct(a, b)
	denom := float32(math.Sqrt(float64(DotProduct(a, a)))) * float32(math.Sqrt(float64(DotProduct(b, b))))
	if denom == 0 {
		return 0
	}
	return dp / denom
}

// EuclideanDistance computes the Euclidean distance between two vectors.
func EuclideanDistance(a, b []float32) float32 {
	if len(a) != len(b) {
		panic("vectors must be of the same length")
	}
	var sum float32
	for i := 0; i < len(a); i++ {
		diff := a[i] - b[i]
		sum += diff * diff
	}
	return float32(math.Sqrt(float64(sum)))
}

// EuclideanSimilarity computes the similarity based on Euclidean distance.
func EuclideanSimilarity(a, b []float32) float32 {
	return 1 / (1 + EuclideanDistance(a, b))
}
