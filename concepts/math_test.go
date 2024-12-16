// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package concepts_test

import (
	"math/rand"
	"testing"
	"tlyakhov/gofoom/concepts"
)

func BenchmarkColorasm(b *testing.B) {
	ca := concepts.Vector4{rand.Float64(), rand.Float64(), rand.Float64(), rand.Float64()}
	ca[0] *= ca[3]
	ca[1] *= ca[3]
	ca[2] *= ca[3]
	//caa := ca
	cb := concepts.Vector4{rand.Float64(), rand.Float64(), rand.Float64(), rand.Float64()}
	cb[0] *= cb[3]
	cb[1] *= cb[3]
	cb[2] *= cb[3]
	o := rand.Float64() * 1.2
	b.Run("Color", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			//			b.StopTimer()

			//b.StartTimer()
			//ca.AddPreMulColorSelfOpacity(&cb, o)
			concepts.BlendColors(&ca, &cb, o)
			//log.Printf("a:%v, b:%v, o: %v, result: %v", caa.StringHuman(), cb.StringHuman(), o, ca.StringHuman())
		}
	})
}

func BlendFrameBufferGo(buffer []uint8, fb []concepts.Vector4, tint *concepts.Vector4) {
	for fbIndex := 0; fbIndex < len(fb); fbIndex++ {
		screenIndex := fbIndex * 4
		inva := 1.0 - tint[3]
		buffer[screenIndex+3] = 0xFF
		buffer[screenIndex+2] = concepts.ByteClamp((fb[fbIndex][2]*inva + tint[2]) * 0xFF)
		buffer[screenIndex+1] = concepts.ByteClamp((fb[fbIndex][1]*inva + tint[1]) * 0xFF)
		buffer[screenIndex+0] = concepts.ByteClamp((fb[fbIndex][0]*inva + tint[0]) * 0xFF)
	}
}

func BenchmarkBlendFrameBuffer(b *testing.B) {
	w := 640
	h := 360
	target := make([]uint8, w*h*4)
	fb := make([]concepts.Vector4, w*h)
	tint := concepts.Vector4{0, 0, 0, 0}
	b.Run("Color", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			concepts.BlendFrameBuffer(target, fb, &tint)
			//BlendFrameBufferGo(target, fb, &tint)
		}
	})
}

func BenchmarkByteClamp(b *testing.B) {
	b.Run("Branchless", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			byteClampBranchless(int(rand.Float64()*300 - 30))
		}
	})
	b.Run("Branching", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			byteClamp(rand.Float64()*300 - 30)
		}
	})
}

func byteClamp(x float64) uint8 {
	if x <= 0 {
		return 0
	}
	if x >= 0xFF {
		return 0xFF
	}

	return uint8(x)
}
func byteClampBranchless(x int) uint8 {
	// if x < 0, x = 0
	z := x & ^(x >> 63)
	return uint8(z ^ ((0xFF ^ z) & ((0xFF - z) >> 63)))
}
