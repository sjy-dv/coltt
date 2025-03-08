// Licensed to sjy-dv under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. sjy-dv licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package distancepq

import (
	"runtime"

	"github.com/rs/zerolog/log"
	"github.com/sjy-dv/coltt/pkg/distancepq/asm"
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
