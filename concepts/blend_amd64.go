//go:build ignore

// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0
package main

import (
	"math"

	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
)

var dataSection Mem
var idxOne, idx255, idxUint32Shuffle, idxReciprocal255 int

func avoidAVXSlowdowns() {
	Comment("This is critical to avoid slowdowns when mixing SSE/AVX code. See")
	Comment("Intel Architectures Optimization Reference Manual Volume 1")
	Comment("section 3.11.6.3 \"Fixing Instruction Sequence Slowdowns\"")
	VZEROUPPER()
}

func genBlendFramebuffer() {

	TEXT("blendFrameBuffer", NOSPLIT, "func(target []uint8, fb [][4]float64, tint *[4]float64)")
	Pragma("noescape")
	Doc("blendFrameBuffer converts float64 framebuffer to uint8 with an added tint")
	fb := Load(Param("fb").Base(), GP64())
	fbPtr := Mem{Base: fb}
	target := Load(Param("target").Base(), GP64())
	targetPtr := Mem{Base: target}
	size := Load(Param("target").Len(), GP64())
	tintPtr := Mem{Base: Load(Param("tint"), GP64())}
	Comment("Set up a zero register")
	zero := YMM()
	VXORPS(zero, zero, zero)
	Comment("Broadcast 1.0 into a register")
	one := YMM()
	VBROADCASTSD(dataSection.Offset(idxOne), one)
	Comment("Broadcast 255.0 into a register")
	y255 := YMM()
	VBROADCASTSD(dataSection.Offset(idx255), y255)
	Comment("tintPacked = *tintPtr")
	tintPacked := YMM()
	VMOVUPD(tintPtr.Offset(0), tintPacked)
	Comment("tintPacked = MIN(tintPacked,1)")
	VMINPD(tintPacked, one, tintPacked)
	Comment("tintPacked = MAX(tintPacked,0)")
	VMAXPD(tintPacked, zero, tintPacked)
	Comment("Put the alpha value tint[3] into a register")
	tintAlpha := YMM()
	VBROADCASTSD(tintPtr.Offset(3*8), tintAlpha)
	Comment("If we have FMA, we could just do VFMADD132PD(b, bAlpha, a)")
	Comment("but I want to target general SSE/AVX")
	Comment("bAlpha = 1.0 - bAlpha")
	VSUBPD(tintAlpha, one, tintAlpha)
	Comment("FB ptr register")
	fbPacked := YMM()
	Label("LoopStart")
	CMPQ(size, Imm(0))
	JE(LabelRef("LoopEnd"))
	inner := func() {
		Comment("fbPacked = *fbPtr")
		VMOVUPD(fbPtr.Offset(0), fbPacked)
		Comment("fbPacked = fbPacked * tintAlpha")
		VMULPD(fbPacked, tintAlpha, fbPacked)
		Comment("fbPacked = fbPacked + tintPacked")
		VADDPD(tintPacked, fbPacked, fbPacked)
		Comment("fbPacked *= 255")
		VMULPD(fbPacked, y255, fbPacked)
		// Hm, do we need this? the CVT/PACK instructions should take care of it
		//Comment("fbPacked = MIN(fbPacked,255)")
		//VMINPD(fbPacked, y255, fbPacked)
		//Comment("fbPacked = MAX(fbPacked,0)")
		//VMAXPD(fbPacked, zero, fbPacked)
		Comment("Convert float64 -> int32")
		targetPacked := XMM()
		VCVTPD2DQY(fbPacked, targetPacked)
		Comment("int32 -> int16")
		VPACKUSDW(targetPacked, targetPacked, targetPacked)
		Comment("int16 -> int8")
		VPACKUSWB(targetPacked, targetPacked, targetPacked)
		VMOVD(targetPacked, targetPtr)
		Comment("Increment pointers")
		ADDQ(Imm(32), fb)
		ADDQ(Imm(4), target)
	}
	// Unroll 4x
	unroll := uint64(4)
	for range unroll {
		inner()
	}
	Comment("Decrement loop counter")
	SUBQ(Imm(4*unroll), size)
	JMP(LabelRef("LoopStart"))
	Label("LoopEnd")

	avoidAVXSlowdowns()
	RET()
}

func genBlendColors() {
	Comment("Benchmarked call overhead is ~2.8 ns/op")
	TEXT("blendColors", NOSPLIT, "func(a, b *[4]float64, opacity float64)")
	Pragma("noescape")
	Doc("blendColors adds a and b * opacity.")
	aptr := Mem{Base: Load(Param("a"), GP64())}
	bptr := Mem{Base: Load(Param("b"), GP64())}
	ob, _ := Param("opacity").Resolve()
	opacity := ob.Addr
	Comment("setup a zero register")
	zero := YMM()
	VXORPS(zero, zero, zero)
	Comment("Broadcast 1.0 into a register")
	one := YMM()
	VBROADCASTSD(dataSection.Offset(idxOne), one)
	Comment("opacityPacked = opacity,opacity,opacity,opacity")
	opacityPacked := YMM()
	VBROADCASTSD(opacity, opacityPacked)
	Comment("Put the alpha value b[3] into a register")
	bAlpha := YMM()
	VBROADCASTSD(bptr.Offset(3*8), bAlpha)
	Comment("bAlpha *= opacity")
	VMULPD(opacityPacked, bAlpha, bAlpha)
	Comment("If we have FMA, we could just do VFMADD132PD(b, bAlpha, a)")
	Comment("but I want to target general SSE/AVX")
	Comment("bAlpha = 1.0 - bAlpha")
	VSUBPD(bAlpha, one, bAlpha)
	Comment("bPacked = *bptr")
	bPacked := YMM()
	VMOVUPD(bptr.Offset(0), bPacked)
	Comment("bPacked = *bptr * opacity")
	VMULPD(opacityPacked, bPacked, bPacked)
	Comment("bPacked = MIN(bPacked,1)")
	VMINPD(bPacked, one, bPacked)
	Comment("bPacked = MAX(bPacked,0)")
	VMAXPD(bPacked, zero, bPacked)
	Comment("aPacked = a")
	aPacked := YMM()
	VMOVUPD(aptr.Offset(0), aPacked)
	Comment("aPacked = aPacked * bAlpha")
	VMULPD(aPacked, bAlpha, aPacked)
	Comment("aPacked = aPacked + bPacked")
	VADDPD(bPacked, aPacked, aPacked)
	Comment("aPacked = MIN(aPacked,1)")
	VMINPD(aPacked, one, aPacked)
	Comment("aPacked = MAX(aPacked,0)")
	VMAXPD(aPacked, zero, aPacked)
	Comment("*aptr = aPacked")
	VMOVUPD(aPacked, aptr.Offset(0))
	avoidAVXSlowdowns()
	RET()
}

func genMul4Self() {
	TEXT("AsmVector4Mul4Self", NOSPLIT, "func(a, b *[4]float64)")
	Pragma("noescape")
	Doc("AsmVector4Mul4Self multiplies a and b.")
	aptr := Mem{Base: Load(Param("a"), GP64())}
	bptr := Mem{Base: Load(Param("b"), GP64())}
	Comment("a = *aptr")
	a := YMM()
	VMOVUPD(aptr.Offset(0), a)
	Comment("a = a * *bptr")
	b := YMM()
	VMOVUPD(bptr.Offset(0), b)
	VMULPD(b, a, a)
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
	Comment("Move quadword from r/m64 to xmm1.")
	VMOVQ(dataSection.Offset(idxUint32Shuffle), shuf)
	Comment("Broadcast 1.0/255.0 into a register")
	reciprocal255 := YMM()
	VBROADCASTSD(dataSection.Offset(idxReciprocal255), reciprocal255)
	Comment("Shuffle bytes in xmm1 according to contents of xmm2/m128.")
	PSHUFB(shuf, c)
	Comment("Zero extend 4 packed 8-bit integers in the low 4 bytes of xmm2/m32 to 4 packed 32-bit integers in xmm")
	PMOVZXBD(c, c)
	Comment("Convert four packed signed doubleword integers from xmm2/mem to four packed double precision floating-point values in ymm1.")
	VCVTDQ2PD(c, a)
	Comment("Multiply packed double precision floating-point values in ymm3/m256 with")
	Comment("ymm2 and store result in ymm1.")
	VMULPD(a, reciprocal255, a)
	Comment("Move unaligned packed double precision floating-point from ymm1 to ymm2/mem.")
	VMOVUPD(a, aptr.Offset(0))
	avoidAVXSlowdowns()
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
	Comment("Move quadword from r/m64 to xmm1.")
	VMOVQ(dataSection.Offset(idxUint32Shuffle), shuf)
	Comment("Broadcast 1.0/255.0 into a register")
	reciprocal255 := YMM()
	VBROADCASTSD(dataSection.Offset(idxReciprocal255), reciprocal255)
	Comment("Shuffle bytes in xmm1 according to contents of xmm2/m128.")
	PSHUFB(shuf, c)
	Comment("Zero extend 4 packed 8-bit integers in the low 4 bytes of xmm2/m32 to 4 packed 32-bit integers in xmm")
	PMOVZXBD(c, c)
	Comment("Convert four packed signed doubleword integers from xmm2/mem to four packed double precision floating-point values in ymm1.")
	VCVTDQ2PD(c, a)
	Comment("Multiply packed double precision floating-point values in ymm3/m256 with")
	Comment("ymm2 and store result in ymm1.")
	VMULPD(a, reciprocal255, a)
	Comment("Move unaligned packed double precision floating-point from ymm1 to ymm2/mem.")
	VMOVUPD(a, aptr.Offset(0))
	avoidAVXSlowdowns()
	RET()
}
func main() {
	dataSection = GLOBL("data", RODATA|NOPTR)
	idxOne = 0
	DATA(0, U64(math.Float64bits(1.0)))
	idx255 = 8
	DATA(8, U64(math.Float64bits(255.0)))
	idxUint32Shuffle = 8
	DATA(16, U64(0b00000000_00000001_00000010_00000011))
	idxReciprocal255 = 16
	DATA(24, U64(math.Float64bits(1.0/255.0)))
	// Benchmarked call overhead is ~2.8 ns/op
	genBlendFramebuffer()
	genBlendColors()
	genMul4Self()
	genInt32ToVector4()
	genInt32ToVector4PreMul()
	Generate()
}
