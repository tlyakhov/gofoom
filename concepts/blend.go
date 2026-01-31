// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package concepts

import (
	"math/rand/v2"
	"unsafe"
)

//go:generate go run blend_amd64.go -out blend_amd64.s -stubs blend_amd64_stub.go

type BlendType int

//go:generate go run github.com/dmarkham/enumer -type=BlendType -json
const (
	BlendNormal BlendType = iota
	BlendDissolve
	BlendMultiply
	BlendScreen
	BlendOverlay
)

// Formulas are from https://en.wikipedia.org/wiki/Blend_modes

var BlendingFuncs = map[BlendType]func(a *Vector4, b *Vector4, opacity float64){
	BlendNormal:   BlendColors,
	BlendDissolve: blendDissolve,
	BlendMultiply: blendMultiply,
	BlendScreen:   blendScreen,
	BlendOverlay:  blendOverlay,
}

// BlendFrameBuffer converts float64 framebuffer to uint8 with an added tint
func BlendFrameBuffer(target []uint8, frameBuffer []Vector4, tint *Vector4) {
	fb := unsafe.Slice((*[4]float64)(unsafe.Pointer(&frameBuffer[0])), len(frameBuffer))
	blendFrameBuffer(target, fb, (*[4]float64)(tint))
}

// BlendColors adds a and b * opacity.
func BlendColors(a, b *Vector4, opacity float64) {
	blendColors((*[4]float64)(a), (*[4]float64)(b), opacity)
}

func blendDissolve(a, b *Vector4, opacity float64) {
	r := rand.Float64()

	if r < b[3] {
		a[2] = b[2] * a[3] / b[3]
		a[1] = b[1] * a[3] / b[3]
		a[0] = b[0] * a[3] / b[3]
	}
}

func blendMultiply(a, b *Vector4, opacity float64) {
	inva := 1.0 - b[3]
	a[0] = a[0] * (inva + b[0])
	a[1] = a[1] * (inva + b[1])
	a[2] = a[2] * (inva + b[2])
	a[3] = a[3]*inva + b[3]
}

func blendScreen(a, b *Vector4, opacity float64) {
	inva := 1.0 - b[3]
	a[0] = 1.0 - (1.0-a[0])*(1.0-b[0])
	a[1] = 1.0 - (1.0-a[1])*(1.0-b[1])
	a[2] = 1.0 - (1.0-a[2])*(1.0-b[2])
	a[3] = a[3]*inva + b[3]
}

func blendOverlay(a, b *Vector4, opacity float64) {
	inva := 1.0 - b[3]
	a[3] = a[3]*inva + b[3]
	if a[0] < 0.5 {
		a[0] = 2.0 * a[0] * (inva + b[0])
	} else {
		a[0] = 1.0 - 2.0*(1.0-a[0])*(1.0-b[0])
	}
	if a[1] < 0.5 {
		a[1] = 2.0 * a[1] * (inva + b[1])
	} else {
		a[1] = 1.0 - 2.0*(1.0-a[1])*(1.0-b[1])
	}
	if a[2] < 0.5 {
		a[2] = 2.0 * a[2] * (inva + b[2])
	} else {
		a[2] = 1.0 - 2.0*(1.0-a[2])*(1.0-b[2])
	}
}
