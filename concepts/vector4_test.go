package concepts

import (
	"testing"
)

func BenchmarkAsmAddPreMulColorSelf(b *testing.B) {
	va := Vector4{1, 0.5, 0.2, 0.5}
	vb := Vector4{0.2, 0.4, 1, 0.25}
	for i := 0; i < b.N; i++ {
		asmAddPreMulColorSelf((*[4]float64)(&va), (*[4]float64)(&vb))
	}
}

func BenchmarkGoAddPreMulColorSelf(b *testing.B) {
	va := Vector4{1, 0.5, 0.2, 0.5}
	vb := Vector4{0.2, 0.4, 1, 0.25}
	for i := 0; i < b.N; i++ {
		va.AddPreMulColorSelf(&vb)
	}
}
