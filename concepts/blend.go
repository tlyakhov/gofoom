// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package concepts

import (
	"math/rand/v2"
	"reflect"
	"unsafe"
)

//go:generate go run blend_amd64.go -out blend_amd64.s -stubs blend_amd64_stub.go

// Formulas are from https://en.wikipedia.org/wiki/Blend_modes
type BlendingFunc func(a, b *Vector4) *Vector4

var BlendingFuncs = map[string]BlendingFunc{
	"Normal":   BlendNormal,
	"Dissolve": BlendDissolve,
	"Multiply": BlendMultiply,
	"Screen":   BlendScreen,
	"Overlay":  BlendOverlay,
}

var BlendingFuncNames map[uintptr]string

func init() {
	BlendingFuncNames = make(map[uintptr]string)
	for name, f := range BlendingFuncs {
		BlendingFuncNames[reflect.ValueOf(f).Pointer()] = name
	}
}

// BlendFrameBuffer converts float64 framebuffer to uint8 with an added tint
func BlendFrameBuffer(target []uint8, frameBuffer []Vector4, tint *Vector4) {
	fb := unsafe.Slice((*[4]float64)(unsafe.Pointer(&frameBuffer[0])), len(frameBuffer))
	blendFrameBuffer(target, fb, (*[4]float64)(tint))
}

// BlendColors adds a and b * opacity.
func BlendColors(a *Vector4, b *Vector4, opacity float64) {
	blendColors((*[4]float64)(a), (*[4]float64)(b), opacity)
}

func BlendNormal(a, b *Vector4) *Vector4 {
	return a.AddPreMulColorSelf(b)
}

func BlendDissolve(a, b *Vector4) *Vector4 {
	r := rand.Float64()

	if r < b[3] {
		a[2] = b[2] * a[3] / b[3]
		a[1] = b[1] * a[3] / b[3]
		a[0] = b[0] * a[3] / b[3]
	}
	return a
}

func BlendMultiply(a, b *Vector4) *Vector4 {
	inva := 1.0 - b[3]
	a[0] = a[0] * (inva + b[0])
	a[1] = a[1] * (inva + b[1])
	a[2] = a[2] * (inva + b[2])
	a[3] = a[3]*inva + b[3]
	return a
}

func BlendScreen(a, b *Vector4) *Vector4 {
	inva := 1.0 - b[3]
	a[0] = 1.0 - (1.0-a[0])*(1.0-b[0])
	a[1] = 1.0 - (1.0-a[1])*(1.0-b[1])
	a[2] = 1.0 - (1.0-a[2])*(1.0-b[2])
	a[3] = a[3]*inva + b[3]
	return a
}

func BlendOverlay(a, b *Vector4) *Vector4 {
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
	return a
}
