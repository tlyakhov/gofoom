// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package materials

import (
	"math"
	"tlyakhov/gofoom/ecs"

	"github.com/spf13/cast"
)

type ToneMap struct {
	ecs.Attached `editable:"^"`

	// 2.4 by default
	Gamma float64 `editable:"Gamma"`

	LutLinearToSRGB [256]float64
	LutSRGBToLinear [256]float64
}

var ToneMapCID ecs.ComponentID

func init() {
	ToneMapCID = ecs.RegisterComponent(&ecs.Column[ToneMap, *ToneMap]{Getter: GetToneMap})
}

func (x *ToneMap) ComponentID() ecs.ComponentID {
	return ToneMapCID
}
func GetToneMap(u *ecs.Universe, e ecs.Entity) *ToneMap {
	panic("Tried to materials.GetToneMap. Use Universe.Singleton(materials.ToneMapCID) instead.")
	/*
		if asserted, ok := u.Component(e, ToneMapCID).(*ToneMap); ok {
			return asserted
		}
		return nil*/
}

func (tm *ToneMap) MultiAttachable() bool { return true }

func (tm *ToneMap) String() string {
	return "ToneMap"
}

func (tm *ToneMap) Recalculate() {
	for i := 0; i < len(tm.LutLinearToSRGB); i++ {
		f := float64(i) / 255.0
		tm.LutLinearToSRGB[i] = tm.LinearTosRGB(f)
		tm.LutSRGBToLinear[i] = tm.SRGBToLinear(f)
	}
}

func (tm *ToneMap) Construct(data map[string]any) {
	tm.Attached.Construct(data)
	tm.Flags |= ecs.ComponentInternal
	tm.Gamma = 2.4
	defer tm.Recalculate()

	if data == nil {
		return
	}

	if v, ok := data["Gamma"]; ok {
		tm.Gamma = cast.ToFloat64(v)
	}
}

const (
	tonemapA = 0.15
	tonemapB = 0.50
	tonemapC = 0.10
	tonemapD = 0.20
	tonemapE = 0.02
	tonemapF = 0.30
	tonemapW = 11.2
)

func tonemap(x float64) float64 {
	return ((x*(tonemapA*x+tonemapC*tonemapB) + tonemapD*tonemapE) / (x*(tonemapA*x+tonemapB) + tonemapD*tonemapF)) - tonemapE/tonemapF
}

func (m *ToneMap) LinearTonemapped(x float64) float64 {
	// Adapted from http://filmicworlds.com/blog/filmic-tonemapping-operators/
	exposureBias := 3.0
	x = tonemap(exposureBias * x)
	whiteScale := 1.0 / tonemap(tonemapW)
	return x*whiteScale - 0.05
}

// Converts a color from linear light gamma to sRGB gamma
func (tm *ToneMap) LinearTosRGB(x float64) float64 {
	// Adapted from https://gamedev.stackexchange.com/questions/92015/optimized-linear-to-srgb-glsl
	if x < 0.0031308 {
		return x * 12.92
	} else {
		return 1.055*math.Pow(x, 1.0/tm.Gamma) - 0.055
	}
}

// Converts a color from sRGB gamma to linear light gamma
func (tm *ToneMap) SRGBToLinear(x float64) float64 {
	// Adapted from https://gamedev.stackexchange.com/questions/92015/optimized-linear-to-srgb-glsl
	if x < 0.04045 {
		return x / 12.92
	} else {
		return math.Pow((x+0.055)/1.055, tm.Gamma)
	}
}
