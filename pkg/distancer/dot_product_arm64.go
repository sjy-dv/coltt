package distancer

import (
	"github.com/sjy-dv/nnv/pkg/distancer/asm"
	"golang.org/x/sys/cpu"
)

func init() {
	if cpu.ARM64.HasASIMD {
		if cpu.ARM64.HasSVE {
			dotProductImplementation = asm.Dot_SVE
		} else {
			dotProductImplementation = asm.Dot_Neon
		}
	}
}
