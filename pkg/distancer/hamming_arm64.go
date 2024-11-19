package distancer

import (
	"github.com/sjy-dv/nnv/pkg/distancer/asm"
	"golang.org/x/sys/cpu"
)

func init() {
	if cpu.ARM64.HasASIMD {
		hammingImpl = asm.Hamming
		hammingBitwiseImpl = asm.HammingBitwise
	}
}
