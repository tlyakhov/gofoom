package state

import (
	"math"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
)

type MaterialSampler struct {
	*Config
	Output concepts.Vector4
	X, Y   int
	Angle  float64
}

func (ms *MaterialSampler) SampleShader(ref *concepts.EntityRef, extraStages []*materials.ShaderStage, u, v float64, scale float64) *concepts.Vector4 {
	ms.Output[0] = 0
	ms.Output[1] = 0
	ms.Output[2] = 0
	ms.Output[3] = 0
	shader := materials.ShaderFromDb(ref)
	if shader == nil {
		ms.sampleTexture(&ms.Output, ref, nil, u, v, scale)
	} else {
		for _, stage := range shader.Stages {
			ms.sampleTexture(&ms.Output, stage.Texture, stage, u, v, scale)
		}
	}

	for _, stage := range extraStages {
		ms.sampleTexture(&ms.Output, stage.Texture, stage, u, v, scale)
	}
	return &ms.Output
}

func (c *MaterialSampler) sampleTexture(result *concepts.Vector4, material *concepts.EntityRef, stage *materials.ShaderStage, u, v float64, scale float64) *concepts.Vector4 {
	// Should refactor this scale thing, it's hard to reason about
	scaleDivisor := 1.0

	if stage != nil {
		u, v = stage.Transform[0]*u+stage.Transform[2]*v+stage.Transform[4], stage.Transform[1]*u+stage.Transform[3]*v+stage.Transform[5]
		if (stage.Flags & materials.ShaderSky) != 0 {
			v = float64(c.Y) / (float64(c.ScreenHeight) - 1)

			if (stage.Flags & materials.ShaderStaticBackground) != 0 {
				u = float64(c.X) / (float64(c.ScreenWidth) - 1)
			} else {
				u = c.Angle / (2.0 * math.Pi)
			}
		}
	}

	if stage == nil || (stage.Flags&materials.ShaderTiled) != 0 {
		u *= scaleDivisor
		v *= scaleDivisor
		if stage != nil && (stage.Flags&materials.ShaderLiquid) != 0 {
			lv, lu := math.Sincos(float64(c.Frame) * constants.LiquidChurnSpeed * concepts.Deg2rad)
			u += lu * constants.LiquidChurnSize
			v += lv * constants.LiquidChurnSize
		}

		u -= math.Floor(u)
		v -= math.Floor(v)
	}

	var sample concepts.Vector4
	if image := materials.ImageFromDb(material); image != nil {
		sample = image.Sample(u, v, scale)
	} else if text := materials.TextFromDb(material); text != nil {
		sample = text.Sample(u, v, scale)
	} else if solid := materials.SolidFromDb(material); solid != nil {
		sample = solid.Diffuse.Render
	} else {
		sample[0] = 0.5
		sample[1] = 0
		sample[2] = 0.5
		sample[3] = 1.0
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
