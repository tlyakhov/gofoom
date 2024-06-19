// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package concepts

import (
	"testing"
)

func BenchmarkAsmAddPreMulColorSelf(b *testing.B) {
	va := Vector4{1, 0.5, 0.2, 0.5}
	vb := Vector4{0.2, 0.4, 1, 0.25}
	for i := 0; i < b.N; i++ {
		AsmVector4AddPreMulColorSelf((*[4]float64)(&va), (*[4]float64)(&vb))
	}
}

func BenchmarkGoAddPreMulColorSelf(b *testing.B) {
	va := Vector4{1, 0.5, 0.2, 0.5}
	vb := Vector4{0.2, 0.4, 1, 0.25}
	for i := 0; i < b.N; i++ {
		va.AddPreMulColorSelf(&vb)
	}
}

func BenchmarkAsmInt32ToVector4(b *testing.B) {
	c := uint32(0x010203FF)
	v := Vector4{}
	for i := 0; i < b.N; i++ {
		AsmInt32ToVector4(c, (*[4]float64)(&v))
	}
}
