package distancer

import (
	"github.com/sjy-dv/nnv/pkg/distancer/asm"
	"golang.org/x/sys/cpu"
)

func init() {
	if cpu.X86.HasAMXBF16 && cpu.X86.HasAVX512 {
		dotProductImplementation = asm.DotAVX512
	} else if cpu.X86.HasAVX2 {
		dotProductImplementation = asm.DotAVX256
	}
}
