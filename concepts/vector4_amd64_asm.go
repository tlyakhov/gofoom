// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

//go:build ignore

package main

import (
	"math"

	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
)

var dataSection Mem
var idxOne, idxUint32Shuffle, idxReciprocal255 int

func genAddSelf() {
	TEXT("AsmVector4AddSelf", NOSPLIT, "func(a, b *[4]float64)")
	Pragma("noescape")
	Doc("AsmVector4AddSelf adds a and b.")
	aptr := Mem{Base: Load(Param("a"), GP64())}
	bptr := Mem{Base: Load(Param("b"), GP64())}
	// a = *aptr
	a := YMM()
	VMOVUPD(aptr.Offset(0), a)
	// a = a + *bptr
	VADDPD(bptr.Offset(0), a, a)
	VMOVUPD(a, aptr.Offset(0))
	RET()
}

func genMul4Self() {
	TEXT("AsmVector4Mul4Self", NOSPLIT, "func(a, b *[4]float64)")
	Pragma("noescape")
	Doc("AsmVector4Mul4Self multiplies a and b.")
	aptr := Mem{Base: Load(Param("a"), GP64())}
	bptr := Mem{Base: Load(Param("b"), GP64())}
	// a = *aptr
	a := YMM()
	VMOVUPD(aptr.Offset(0), a)
	// a = a * *bptr
	b := YMM()
	VMOVUPD(bptr.Offset(0), b)
	VMULPD(b, a, a)
	VMOVUPD(a, aptr.Offset(0))
	RET()
}

func genAddPreMulColorSelf() {
	TEXT("AsmVector4AddPreMulColorSelf", NOSPLIT, "func(a, b *[4]float64)")
	Pragma("noescape")
	Doc("AsmVector4AddPreMulColorSelf adds a and b with pre-multiplied alpha.")
	aptr := Mem{Base: Load(Param("a"), GP64())}
	bptr := Mem{Base: Load(Param("b"), GP64())}
	// Put the alpha value b[3] into a register
	bAlpha := YMM()
	VBROADCASTSD(bptr.Offset(3*8), bAlpha)
	// Put 1.0 into a register
	one := YMM()
	VBROADCASTSD(dataSection.Offset(idxOne), one)
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
}

func genInt32ToVector4() {
	TEXT("AsmInt32ToVector4", NOSPLIT, "func(c uint32, a *[4]float64)")
	Pragma("noescape")
	Doc("AsmInt32ToVector4 converts a uint32 color to a vector.")
	c := Load(Param("c"), XMM())
	aptr := Mem{Base: Load(Param("a"), GP64())}
	a := YMM()
	shuf := XMM()
	// Move quadword from r/m64 to xmm1.
	VMOVQ(dataSection.Offset(idxUint32Shuffle), shuf)
	// Broadcast 1.0/255.0 into a register
	reciprocal255 := YMM()
	VBROADCASTSD(dataSection.Offset(idxReciprocal255), reciprocal255)
	// Shuffle bytes in xmm1 according to contents of xmm2/m128.
	PSHUFB(shuf, c)
	// Zero extend 4 packed 8-bit integers in the low 4 bytes of xmm2/m32 to 4 packed 32-bit integers in xmm
	PMOVZXBD(c, c)
	// Convert four packed signed doubleword integers from xmm2/mem to four packed double precision floating-point values in ymm1.
	VCVTDQ2PD(c, a)
	// Multiply packed double precision floating-point values in ymm3/m256 with
	// ymm2 and store result in ymm1.
	VMULPD(a, reciprocal255, a)
	// Move unaligned packed double precision floating-point from ymm1 to ymm2/mem.
	VMOVUPD(a, aptr.Offset(0))
	RET()
}

func genInt32ToVector4PreMul() {
	TEXT("AsmInt32ToVector4PreMul", NOSPLIT, "func(c uint32, a *[4]float64)")
	Pragma("noescape")
	Doc("AsmInt32ToVector4PreMul converts a uint32 color to a vector with pre-multiplied alpha.")
	c := Load(Param("c"), XMM())
	aptr := Mem{Base: Load(Param("a"), GP64())}
	a := YMM()
	shuf := XMM()
	// Move quadword from r/m64 to xmm1.
	VMOVQ(dataSection.Offset(idxUint32Shuffle), shuf)
	// Broadcast 1.0/255.0 into a register
	reciprocal255 := YMM()
	VBROADCASTSD(dataSection.Offset(idxReciprocal255), reciprocal255)
	// Shuffle bytes in xmm1 according to contents of xmm2/m128.
	PSHUFB(shuf, c)
	// Zero extend 4 packed 8-bit integers in the low 4 bytes of xmm2/m32 to 4 packed 32-bit integers in xmm
	PMOVZXBD(c, c)
	// Convert four packed signed doubleword integers from xmm2/mem to four packed double precision floating-point values in ymm1.
	VCVTDQ2PD(c, a)
	// Multiply packed double precision floating-point values in ymm3/m256 with
	// ymm2 and store result in ymm1.
	VMULPD(a, reciprocal255, a)
	// Move unaligned packed double precision floating-point from ymm1 to ymm2/mem.
	VMOVUPD(a, aptr.Offset(0))
	RET()
}

func main() {
	dataSection = GLOBL("data", RODATA|NOPTR)
	// 1.0
	idxOne = 0
	DATA(0, U64(math.Float64bits(1.0)))
	idxUint32Shuffle = 8
	DATA(8, U64(0b00000000_00000001_00000010_00000011))
	idxReciprocal255 = 16
	DATA(16, U64(math.Float64bits(1.0/255.0)))
	genAddSelf()
	genAddPreMulColorSelf()
	genMul4Self()
	genInt32ToVector4()
	genInt32ToVector4PreMul()

	Generate()
}
