package distance

import (
	"runtime"

	"github.com/rs/zerolog/log"
	"github.com/sjy-dv/nnv/pkg/distance/asm"
	"golang.org/x/sys/cpu"
)

func init() {
	if cpu.X86.HasAVX2 && cpu.X86.HasFMA && cpu.X86.HasSSE3 {
		log.Info().Str("GOARCH", runtime.GOARCH).Msg("Using ASM support for dot and euclidean distance")
		dotProductImpl = asm.Dot
		euclideanDistance = asm.SquaredEuclideanDistance
	} else {
		log.Warn().Str("GOARCH", runtime.GOARCH).Msg("No ASM support for dot and euclidean distance")
	}
}
