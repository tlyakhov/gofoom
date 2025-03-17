// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package render

import (
	"math"
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/components/selection"
	"tlyakhov/gofoom/concepts"
)

func (r *Renderer) renderBody(ebd *entityWithDist2, block *block, xStart, xEnd int) {
	b := ebd.Body
	// If it's lit, render a pixel
	if ebd.Visible.PixelOnly {
		r.renderBodyPixel(ebd, block, xStart, xEnd)
		return
	}

	if ebd.Visible.Opacity <= 0 {
		return
	}

	// Calculate angles for picking the right sprite, and also relative to camera
	angleFromPlayer := r.PlayerBody.Angle2DTo(b.Pos.Render)
	// 0 degrees is facing right. Why do we need to adjust by 90 degrees?
	block.SpriteAngle = concepts.NormalizeAngle(*b.Angle.Render - angleFromPlayer + 270)
	angleRender := concepts.NormalizeAngle(angleFromPlayer - *r.PlayerBody.Angle.Render)
	if angleRender >= 180.0 {
		angleRender -= 360.0
	}
	// Calculate screenspace coordinates
	block.Distance = math.Sqrt(ebd.Dist2)
	xMid := math.Tan(angleRender*concepts.Deg2rad)*r.CameraToProjectionPlane + float64(r.ScreenWidth)*0.5
	depthScale := r.CameraToProjectionPlane / math.Cos(angleRender*concepts.Deg2rad)
	depthScale /= block.Distance
	xScale := depthScale * b.Size.Render[0]
	block.ScaleW = uint32(xScale)
	x1 := concepts.Max(int(xMid-xScale*0.5), xStart)
	x2 := concepts.Min(int(xMid+xScale*0.5), xEnd)
	if x1 == x2 || x2 < xStart || x1 >= xEnd {
		return
	}

	block.ProjectedTop = (b.Pos.Render[2] + b.Size.Render[1]*0.5 - block.CameraZ) * depthScale
	block.ProjectedBottom = block.ProjectedTop - b.Size.Render[1]*depthScale
	screenTop := block.ScreenHeight/2 - int(math.Floor(block.ProjectedTop))
	screenBottom := block.ScreenHeight/2 - int(math.Floor(block.ProjectedBottom))
	block.ScaleH = uint32(screenBottom - screenTop)
	block.ClippedTop = concepts.Clamp(screenTop, 0, r.ScreenHeight)
	block.ClippedBottom = concepts.Clamp(screenBottom, 0, r.ScreenHeight)

	if block.Pick &&
		block.ScreenX >= x1 && block.ScreenX < x2 &&
		block.ScreenY >= block.ClippedTop && block.ScreenY <= block.ClippedBottom {

		block.PickedSelection = append(block.PickedSelection, selection.SelectableFromBody(b))
		return
	}

	if lit := materials.GetLit(r.Universe, b.Entity); lit != nil {
		ls := &block.LightSampler
		ls.Sector = b.Sector()
		ls.Normal[0] = math.Cos(*b.Angle.Render * concepts.Deg2rad)
		ls.Normal[1] = math.Sin(*b.Angle.Render * concepts.Deg2rad)
		ls.Normal[2] = 0
		ls.Hash = block.WorldToLightmapHash(ls.Sector, b.Pos.Render, &ls.Normal)
		ls.Segment = nil
		ls.InputBody = b.Entity
		ls.Get()
		block.Light[0] = ls.Output[0]
		block.Light[1] = ls.Output[1]
		block.Light[2] = ls.Output[2]
		block.Light[3] = 1
		// result = Surface * Diffuse * (Ambient + Lightmap)
		block.Light.To3D().AddSelf(&lit.Ambient)
		block.Light.Mul4Self(&lit.Diffuse)
	} else {
		block.Light[0] = 1
		block.Light[1] = 1
		block.Light[2] = 1
		block.Light[3] = 1
	}
	if alive := behaviors.GetAlive(r.Universe, b.Entity); alive != nil {
		alive.Tint(&block.Light)
	}

	vStart := float64(block.ScreenHeight/2) - block.ProjectedTop
	block.Light.MulSelf(ebd.Visible.Opacity)
	block.MaterialSampler.Initialize(b.Entity, nil)
	for block.ScreenX = x1; block.ScreenX < x2; block.ScreenX++ {
		if block.ScreenX < xStart || block.ScreenX >= xEnd {
			continue
		}
		// TODO: Need to double-check the math here, this formula seems iffy at
		// being texel/pixel-exact
		block.MaterialSampler.NU = 0.5 + (float64(block.ScreenX)+1-xMid)/xScale

		for y := block.ClippedTop; y < block.ClippedBottom; y++ {
			screenIndex := (y*r.ScreenWidth + block.ScreenX)
			if block.Distance >= r.ZBuffer[screenIndex] {
				continue
			}
			block.NV = (float64(y) - vStart) / (block.ProjectedTop - block.ProjectedBottom)
			block.MaterialSampler.U = block.NU
			block.MaterialSampler.V = block.NV
			block.SampleMaterial(nil)
			block.MaterialSampler.Output.Mul4Self(&block.Light)
			concepts.BlendColors(&r.FrameBuffer[screenIndex], &block.MaterialSampler.Output, 1.0)
			if block.MaterialSampler.Output[3] > 0.8 {
				r.ZBuffer[screenIndex] = block.Distance
			}
		}
	}
}

func (r *Renderer) renderBodyPixel(ebd *entityWithDist2, block *block, sx, ex int) {
	b := ebd.Body
	lit := materials.GetLit(r.Universe, b.Entity)
	scr := r.WorldToScreen(b.Pos.Render)
	if scr == nil || lit == nil {
		return
	}
	x := int(scr[0])
	y := int(scr[1])
	if x < sx || x >= ex || y < 0 || y >= r.ScreenHeight {
		return
	}
	dist := math.Sqrt(ebd.Dist2)
	screenIndex := x + y*r.ScreenWidth
	if dist >= r.ZBuffer[screenIndex] {
		return
	}

	/*le := &c.LightSampler
	le.Q.From(b.Pos.Render)
	le.MapIndex = c.WorldToLightmapAddress(b.Sector(), &le.Q, 0)
	le.Segment = nil
	le.Type = state.LightSamplerBody
	le.InputBody = b.Entity
	le.Sector = b.Sector()
	le.Get()*/
	block.Light.From(&lit.Diffuse)
	block.Light.To3D().AddSelf(&lit.Ambient)
	concepts.BlendColors(&r.FrameBuffer[screenIndex], &block.Light, 1.0)
	if block.Light[3] > 0.8 {
		r.ZBuffer[screenIndex] = dist
	}
}
