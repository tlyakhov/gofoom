//go:build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
)

func genAddSelf() {
	TEXT("asmVector3AddSelf", NOSPLIT, "func(a, b *[4]float64)")
	Pragma("noescape")
	Doc("asmVector3AddSelf adds a and b.")
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
func genSubSelf() {
	TEXT("asmVector3SubSelf", NOSPLIT, "func(a, b *[4]float64)")
	Pragma("noescape")
	Doc("asmVector3SubSelf subtracts a and b.")
	aptr := Mem{Base: Load(Param("a"), GP64())}
	bptr := Mem{Base: Load(Param("b"), GP64())}
	// a = *aptr
	a := YMM()
	VMOVUPD(aptr.Offset(0), a)
	// a = a - *bptr
	VSUBPD(bptr.Offset(0), a, a)
	VMOVUPD(a, aptr.Offset(0))
	RET()
}

func main() {
	genAddSelf()
	genSubSelf()

	Generate()
}
