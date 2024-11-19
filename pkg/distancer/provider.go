package distancer

type Provider interface {
	New(vec []float32) Distancer
	SingleDist(vec1, vec2 []float32) (float32, error)
	Step(x, y []float32) float32
	Wrap(x float32) float32
	Type() string
}

type Distancer interface {
	Distance(vec []float32) (float32, error)
}
