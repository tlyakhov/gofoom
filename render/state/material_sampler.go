// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package state

import (
	"math"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/dynamic"
	"tlyakhov/gofoom/ecs"
)

type MaterialSampler struct {
	*Config
	*Ray
	Output           concepts.Vector4
	ScreenX, ScreenY int
	ScaleW, ScaleH   uint32
	NoTexture        bool
	SpriteAngle      float64
	Materials        []ecs.Attachable
	pipelineIndex    int
	U, V             float64
	NU, NV           float64
}

func (ms *MaterialSampler) Initialize(material ecs.Entity, extraStages []*materials.ShaderStage) {
	ms.Materials = ms.Materials[:0]
	ms.derefMaterials(material, nil)
	for _, stage := range extraStages {
		ms.derefMaterials(stage.Texture, nil)
	}
}

func (ms *MaterialSampler) InitializeRayBody(src, dst *concepts.Vector3, b *core.Body) bool {
	delta := &concepts.Vector3{dst[0] - src[0], dst[1] - src[1], dst[2] - src[2]}
	delta.NormSelf()
	isect := concepts.Vector3{}
	seg := b.BillboardSegment(delta, dynamic.DynamicRender)
	ok := seg.Intersect3D(src, dst, &isect)
	if !ok {
		return false
	}
	ms.Initialize(b.Entity, nil)
	ms.NU = isect.To2D().Dist(seg.A) / b.Size.Render[0]
	ms.NV = (b.Pos.Render[2] + b.Size.Render[0]*0.5 - isect[2]) / (b.Size.Render[1])
	if ms.NV < 0 || ms.NV > 1 {
		return false
	}
	ms.U = ms.NU
	ms.V = ms.NV
	return true
}

func (ms *MaterialSampler) derefMaterials(material ecs.Entity, parent ecs.Attachable) {
	if shader := materials.GetShader(ms.ECS, material); shader != nil && shader != parent {
		ms.Materials = append(ms.Materials, shader)
		for _, stage := range shader.Stages {
			ms.derefMaterials(stage.Texture, shader)
		}
	} else if sprite := materials.GetSprite(ms.ECS, material); sprite != nil && sprite != parent {
		ms.Materials = append(ms.Materials, sprite)
		ms.derefMaterials(sprite.Image, sprite)
	} else if image := materials.GetImage(ms.ECS, material); image != nil {
		ms.Materials = append(ms.Materials, image)
	} else if text := materials.GetText(ms.ECS, material); text != nil {
		ms.Materials = append(ms.Materials, text)
	} else if solid := materials.GetSolid(ms.ECS, material); solid != nil {
		ms.Materials = append(ms.Materials, solid)
	}

}

func (ms *MaterialSampler) SampleMaterial(extraStages []*materials.ShaderStage) {
	ms.Output[0] = 0
	ms.Output[1] = 0
	ms.Output[2] = 0
	ms.Output[3] = 0
	ms.pipelineIndex = 0
	ms.sampleStage(nil)
	for _, stage := range extraStages {
		ms.sampleStage(stage)
	}
}

func (ms *MaterialSampler) sampleStage(stage *materials.ShaderStage) {
	u := ms.U
	v := ms.V
	if stage != nil {
		if stage.IgnoreSurfaceTransform {
			u, v = stage.Transform[0]*ms.NU+stage.Transform[2]*ms.NV+stage.Transform[4], stage.Transform[1]*ms.NU+stage.Transform[3]*ms.NV+stage.Transform[5]
		} else {
			u, v = stage.Transform[0]*u+stage.Transform[2]*v+stage.Transform[4], stage.Transform[1]*u+stage.Transform[3]*v+stage.Transform[5]
		}
		if (stage.Flags & materials.ShaderSky) != 0 {
			v = float64(ms.ScreenY) / (float64(ms.ScreenHeight) - 1)

			if (stage.Flags & materials.ShaderStaticBackground) != 0 {
				u = float64(ms.ScreenX) / (float64(ms.ScreenWidth) - 1)
			} else {
				u = ms.Angle / (2.0 * math.Pi)
			}
		}

		if (stage.Flags & materials.ShaderLiquid) != 0 {
			lv, lu := math.Sincos(float64(stage.ECS.Frame) * constants.LiquidChurnSpeed * concepts.Deg2rad)
			u += lu * constants.LiquidChurnSize
			v += lv * constants.LiquidChurnSize
		}

	}

	if stage == nil || (stage.Flags&materials.ShaderTiled) != 0 {
		u -= math.Floor(u)
		v -= math.Floor(v)
	}

	ms.NoTexture = false
	var a ecs.Attachable
	if ms.pipelineIndex < len(ms.Materials) {
		a = ms.Materials[ms.pipelineIndex]
	}
	switch m := a.(type) {
	case *materials.Shader:
		ms.pipelineIndex++
		for _, stage := range m.Stages {
			ms.sampleStage(stage)
		}
	case *materials.Sprite:
		ms.pipelineIndex++
		frame := uint32(*m.Frame.Render)
		if stage != nil {
			frame += uint32(stage.Frame)
		}
		cell := uint32(ms.SpriteAngle) * m.Angles / 360
		cell += m.Angles * (frame % m.Frames)

		c := cell % m.Cols
		r := cell / m.Cols
		ms.U, ms.V = m.TransformUV(u, v, c, r)
		ms.ScaleW *= m.Cols
		ms.ScaleH *= m.Rows
		ms.sampleStage(nil)
		ms.ScaleW /= m.Cols
		ms.ScaleH /= m.Rows
	case *materials.Image:
		ms.pipelineIndex++
		sample := m.Sample(u, v, ms.ScaleW, ms.ScaleH)
		ms.Output.AddPreMulColorSelf(&sample)
	case *materials.Text:
		ms.pipelineIndex++
		sample := m.Sample(u, v, ms.ScaleW, ms.ScaleH)
		ms.Output.AddPreMulColorSelf(&sample)
	case *materials.Solid:
		ms.pipelineIndex++
		ms.Output.AddPreMulColorSelf(m.Diffuse.Render)
	default:
		ms.pipelineIndex++
		sample := concepts.Vector4{0.5, 0, 0.5, 1}
		ms.NoTexture = true
		ms.Output.AddPreMulColorSelf(&sample)
	}
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
