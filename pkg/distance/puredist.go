package distance

func squaredEuclideanDistancePureGo(x, y []float32) float32 {
	var sum float32
	for i := range x {
		diff := x[i] - y[i]
		sum += diff * diff
	}
	return sum
}

func dotProductPureGo(x, y []float32) float32 {
	var sum float32
	for i := range x {
		sum += x[i] * y[i]
	}
	return sum
}
