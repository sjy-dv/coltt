//go:generate go run euclidean.go -out ../euclidean.s -stubs ../euclidean_stub.go -pkg asm

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/reg"
)

// Squared Euclidean Distance

// Credit: https://github.com/mmcloughlin/avo/tree/master/examples/dot
/* The dot product code is adjusted so that the difference is computed first,
 * then the rest remains the same. i.e. squared euclidean distance between x and
 * y is: (x-y)(x-y)^T */

var unroll = 4

func main() {
	TEXT("SquaredEuclideanDistance", NOSPLIT, "func(x, y []float32) float32")
	x := Mem{Base: Load(Param("x").Base(), GP64())}
	y := Mem{Base: Load(Param("y").Base(), GP64())}
	n := Load(Param("x").Len(), GP64())

	// Allocate accumulation registers.
	acc := make([]VecVirtual, unroll)
	diff := make([]VecVirtual, unroll)
	for i := 0; i < unroll; i++ {
		acc[i] = YMM()
		diff[i] = YMM()
	}

	// Zero initialization.
	for i := 0; i < unroll; i++ {
		VXORPS(acc[i], acc[i], acc[i])
		VXORPS(diff[i], diff[i], diff[i])
	}

	// Loop over blocks and process them with vector instructions.
	blockitems := 8 * unroll
	blocksize := 4 * blockitems
	Label("blockloop")
	CMPQ(n, U32(blockitems))
	JL(LabelRef("tail"))

	// Load x.
	xs := make([]VecVirtual, unroll)
	for i := 0; i < unroll; i++ {
		xs[i] = YMM()
	}

	for i := 0; i < unroll; i++ {
		VMOVUPS(x.Offset(32*i), xs[i])
	}

	// The actual FMA.
	for i := 0; i < unroll; i++ {
		// Compute difference first
		VSUBPS(y.Offset(32*i), xs[i], diff[i])
		// Then compute the dot product with itself
		VFMADD231PS(diff[i], diff[i], acc[i])
	}

	ADDQ(U32(blocksize), x.Base)
	ADDQ(U32(blocksize), y.Base)
	SUBQ(U32(blockitems), n)
	JMP(LabelRef("blockloop"))

	// Process any trailing entries.
	Label("tail")
	tail := XMM()
	tailDiff := XMM()
	VXORPS(tail, tail, tail)
	VXORPS(tailDiff, tailDiff, tailDiff)

	Label("tailloop")
	CMPQ(n, U32(0))
	JE(LabelRef("reduce"))

	xt := XMM()
	VMOVSS(x, xt)
	// Same idea here, compute difference first, then dot product with itself
	VSUBSS(y, xt, tailDiff)
	VFMADD231SS(tailDiff, tailDiff, tail)

	ADDQ(U32(4), x.Base)
	ADDQ(U32(4), y.Base)
	DECQ(n)
	JMP(LabelRef("tailloop"))

	// Reduce the lanes to one.
	Label("reduce")
	for i := 1; i < unroll; i++ {
		VADDPS(acc[0], acc[i], acc[0])
	}

	result := acc[0].AsX()
	top := XMM()
	VEXTRACTF128(U8(1), acc[0], top)
	VADDPS(result, top, result)
	VADDPS(result, tail, result)
	VHADDPS(result, result, result)
	VHADDPS(result, result, result)
	Store(result, ReturnIndex(0))

	RET()

	Generate()
}
