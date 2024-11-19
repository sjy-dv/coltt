package distancer

import (
	"github.com/sjy-dv/nnv/pkg/distancer/asm"
	"golang.org/x/sys/cpu"
)

func init() {
	if cpu.X86.HasAMXBF16 && cpu.X86.HasAVX512 {
		hammingImpl = asm.HammingAVX512
	} else if cpu.X86.HasAVX2 {
		hammingImpl = asm.HammingAVX256
	}
	if cpu.X86.HasAVX2 {
		hammingBitwiseImpl = asm.HammingBitwiseAVX256
	}
}
