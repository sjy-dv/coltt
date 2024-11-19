package distancer

import (
	"github.com/sjy-dv/nnv/pkg/distancer/asm"
	"golang.org/x/sys/cpu"
)

func init() {
	if cpu.ARM64.HasASIMD {
		if cpu.ARM64.HasSVE {
			l2SquaredImpl = asm.L2_SVE
		} else {
			l2SquaredImpl = asm.L2_Neon
		}
	}
}
