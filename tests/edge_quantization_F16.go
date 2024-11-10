// edge_quantization_f16.go
package edge

import (
	"encoding/binary"

	"github.com/sjy-dv/nnv/pkg/float16"
	"github.com/viterin/vek/vek32"
)

type float16Vec []float16.Float16

var _ Quantization[float16Vec] = Float16Quantization{}

type Float16Quantization struct {
	bufx, bufy Vector
}

func (q Float16Quantization) Similarity(x, y float16Vec) float32 {
	if q.bufx == nil {
		q.bufx = make(Vector, len(x))
		q.bufy = make(Vector, len(x))
	}
	for i := range x {
		q.bufx[i] = x[i].Float32()
		q.bufy[i] = y[i].Float32()
	}
	return vek32.CosineSimilarity(q.bufx, q.bufy)
}

func (q Float16Quantization) Lower(v Vector) (float16Vec, error) {
	out := make(float16Vec, len(v))
	for i, x := range v {
		out[i] = float16.Fromfloat32(x)
	}
	return out, nil
}

func (q Float16Quantization) Marshal(to []byte, lower float16Vec) error {
	for i, n := range lower {
		u := n.Bits()
		binary.LittleEndian.PutUint16(to[i*2:], u)
	}
	return nil
}

func (q Float16Quantization) Unmarshal(data []byte) (float16Vec, error) {
	out := make(float16Vec, len(data)>>1)
	for i := 0; i < len(data); i += 2 {
		bits := binary.LittleEndian.Uint16(data[i:])
		out[i>>1] = float16.Frombits(bits)
	}
	return out, nil
}

func (q Float16Quantization) Name() string {
	return "float16"
}

func (q Float16Quantization) LowerSize(dim int) int {
	return 2 * dim
}
