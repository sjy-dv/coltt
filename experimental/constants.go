package experimental

import "math"

type Vector []float32

func (v Vector) Dimensions() int {
	return len(v)
}

type VertexEdge struct {
	MultiVectors map[string]Vector
	Metadata     map[string]any
}

var (
	panicr = "panic %v"
)

const (
	COSINE                 = "cosine"
	EUCLIDEAN              = "euclidean"
	NONE_QAUNTIZATION      = "none"
	F16_QUANTIZATION       = "f16"
	F8_QUANTIZATION        = "f8"
	BF16_QUANTIZATION      = "bf16"
	T_COSINE               = "cosine-dot"
	VERTEX_SHARD_COUNT int = 16
)

func Normalize(v []float32) []float32 {
	var norm float32
	out := make([]float32, len(v))
	for i := range v {
		norm += v[i] * v[i]
	}
	if norm == 0 {
		return out
	}

	norm = float32(math.Sqrt(float64(norm)))
	for i := range v {
		out[i] = v[i] / norm
	}

	return out
}
