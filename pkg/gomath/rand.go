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
