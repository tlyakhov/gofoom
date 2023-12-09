//go:build ignore

package main

import (
	"math"

	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
)

func main() {
	TEXT("asmAddPreMulColorSelf", NOSPLIT, "func(a, b *[4]float64)")
	Doc("asmAddPreMulColorSelf adds a and b with pre-multiplied alpha.")
	aptr := Mem{Base: Load(Param("a"), GP64())}
	bptr := Mem{Base: Load(Param("b"), GP64())}
	// Put the alpha value b[3] into a register
	bAlpha := YMM()
	VBROADCASTSD(bptr.Offset(3*8), bAlpha)
	// Put 1.0 into a register
	one := YMM()
	VBROADCASTSD(ConstData("one", U64(math.Float64bits(1.0))), one)
	// bAlpha = 1.0 - b[3]
	VSUBPD(bAlpha, one, bAlpha)
	// If we have FMA, we could just do VFMADD132PD(b, bAlpha, a)
	// but I want to target general SSE/AVX
	// a = *aptr * (1.0 - b[3])
	a := YMM()
	VMULPD(aptr.Offset(0), bAlpha, a)
	// a = a + *bptr
	VADDPD(bptr.Offset(0), a, a)
	VMOVUPD(a, aptr.Offset(0))
	RET()
	Generate()
}
