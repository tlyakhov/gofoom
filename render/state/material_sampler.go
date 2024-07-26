// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package state

import (
	"math"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
)

type MaterialSampler struct {
	*Config
	*Ray
	Output           concepts.Vector4
	ScreenX, ScreenY int
	NoTexture        bool
}

func (ms *MaterialSampler) SampleShader(eShader concepts.Entity, extraStages []*materials.ShaderStage, u, v float64, sw, sh uint32) *concepts.Vector4 {
	ms.Output[0] = 0
	ms.Output[1] = 0
	ms.Output[2] = 0
	ms.Output[3] = 0
	shader := materials.ShaderFromDb(ms.DB, eShader)
	if shader == nil {
		ms.sampleTexture(&ms.Output, eShader, nil, u, v, sw, sh)
	} else {
		for _, stage := range shader.Stages {
			ms.sampleTexture(&ms.Output, stage.Texture, stage, u, v, sw, sh)
		}
	}

	for _, stage := range extraStages {
		ms.sampleTexture(&ms.Output, stage.Texture, stage, u, v, sw, sh)
	}
	return &ms.Output
}

func (ms *MaterialSampler) sampleTexture(result *concepts.Vector4, material concepts.Entity, stage *materials.ShaderStage, u, v float64, sw, sh uint32) *concepts.Vector4 {
	if stage != nil {
		u, v = stage.Transform[0]*u+stage.Transform[2]*v+stage.Transform[4], stage.Transform[1]*u+stage.Transform[3]*v+stage.Transform[5]
		if (stage.Flags & materials.ShaderSky) != 0 {
			v = float64(ms.ScreenY) / (float64(ms.ScreenHeight) - 1)

			if (stage.Flags & materials.ShaderStaticBackground) != 0 {
				u = float64(ms.ScreenX) / (float64(ms.ScreenWidth) - 1)
			} else {
				u = ms.Angle / (2.0 * math.Pi)
			}
		}

		if (stage.Flags & materials.ShaderLiquid) != 0 {
			lv, lu := math.Sincos(float64(ms.Frame) * constants.LiquidChurnSpeed * concepts.Deg2rad)
			u += lu * constants.LiquidChurnSize
			v += lv * constants.LiquidChurnSize
		}

	}

	if stage == nil || (stage.Flags&materials.ShaderTiled) != 0 {
		u -= math.Floor(u)
		v -= math.Floor(v)
	}

	ms.NoTexture = false
	var sample concepts.Vector4
	if image := materials.ImageFromDb(ms.DB, material); image != nil {
		sample = image.Sample(u, v, sw, sh)
	} else if text := materials.TextFromDb(ms.DB, material); text != nil {
		sample = text.Sample(u, v, sw, sh)
	} else if solid := materials.SolidFromDb(ms.DB, material); solid != nil {
		sample.From(solid.Diffuse.Render)
	} else {
		sample[0] = 0.5
		sample[1] = 0
		sample[2] = 0.5
		sample[3] = 1.0
		ms.NoTexture = true
	}
	result.AddPreMulColorSelf(&sample)

	return result
}

func WeightBlendedOIT(c *concepts.Vector4, z float64) float64 {
	w := c[0]
	if c[1] > c[0] {
		w = c[1]
	}
	if c[2] > c[0] {
		w = c[2]
	}
	w = concepts.Clamp(w*c[3], c[3], 1.0)
	w *= concepts.Clamp(0.03/(1e-5+math.Pow(z/500, 4.0)), 1e-2, 3e3)
	return w
}
