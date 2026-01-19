// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

//go:build !amd64

package concepts

// blendFrameBuffer converts float64 framebuffer to uint8 with an added tint
func blendFrameBuffer(target []uint8, fb [][4]float64, tint *[4]float64) {
	if tint[3] != 0 {
		for fbIndex := 0; fbIndex < len(fb); fbIndex++ {
			screenIndex := fbIndex * 4
			inva := 1.0 - tint[3]
			target[screenIndex+3] = 0xFF
			target[screenIndex+2] = ByteClamp((fb[fbIndex][2]*inva + tint[2]) * 0xFF)
			target[screenIndex+1] = ByteClamp((fb[fbIndex][1]*inva + tint[1]) * 0xFF)
			target[screenIndex+0] = ByteClamp((fb[fbIndex][0]*inva + tint[0]) * 0xFF)
		}
	} else {
		for fbIndex := 0; fbIndex < len(fb); fbIndex++ {
			screenIndex := fbIndex * 4
			target[screenIndex+3] = 0xFF
			target[screenIndex+2] = ByteClamp(fb[fbIndex][2] * 0xFF)
			target[screenIndex+1] = ByteClamp(fb[fbIndex][1] * 0xFF)
			target[screenIndex+0] = ByteClamp(fb[fbIndex][0] * 0xFF)
		}
	}
}

// blendColors adds a and b * opacity.
func blendColors(a *[4]float64, b *[4]float64, o float64) {
	if b[3] == 0 {
		return
	}
	if b[3] == 1 && o == 1 {
		*a = *b
		return
	}
	inva := 1.0 - b[3]*o
	a[0] = a[0]*inva + b[0]*o
	a[1] = a[1]*inva + b[1]*o
	a[2] = a[2]*inva + b[2]*o
	a[3] = a[3]*inva + b[3]*o
	a[0] = Clamp(a[0], 0, 1)
	a[1] = Clamp(a[1], 0, 1)
	a[2] = Clamp(a[2], 0, 1)
	a[3] = Clamp(a[3], 0, 1)
}

func AsmVector4Mul4Self(a *[4]float64, b *[4]float64) {}

// AsmInt32ToVector4 converts a uint32 color to a vector.
func AsmInt32ToVector4(c uint32, a *[4]float64) {
	//return uint32(v[0]*255)<<24 | uint32(v[1]*255)<<16 | uint32(v[2]*255)<<8 | uint32(v[3]*255)
	a[0] = float64(c>>24) / 255.0
	a[1] = float64((c>>16)&0xFF) / 255.0
	a[2] = float64((c>>8)&0xFF) / 255.0
	a[3] = float64(c&0xFF) / 255.0
}

// AsmInt32ToVector4PreMul converts a uint32 color to a vector with pre-multiplied alpha.
func AsmInt32ToVector4PreMul(c uint32, a *[4]float64) {}
