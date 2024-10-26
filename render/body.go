// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package render

import (
	"math"
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/components/selection"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/render/state"
)

func (r *Renderer) renderBody(ebd *state.EntityWithDist2, c *state.Column, xStart, xEnd int) {
	b := ebd.Body
	// If it's lit, render a pixel
	if ebd.Visible.PixelOnly {
		r.renderBodyPixel(ebd, c, xStart, xEnd)
		return
	}

	if ebd.Visible.Opacity <= 0 {
		return
	}

	// Calculate angles for picking the right sprite, and also relative to camera
	angleFromPlayer := r.PlayerBody.Angle2DTo(b.Pos.Render)
	// This formula seems magic, what's happening here?
	c.SpriteAngle = 270 - angleFromPlayer + *b.Angle.Render + 360
	for c.SpriteAngle > 360 {
		c.SpriteAngle -= 360
	}
	angleRender := angleFromPlayer - *r.PlayerBody.Angle.Render
	for angleRender < -180.0 {
		angleRender += 360.0
	}
	for angleRender > 180.0 {
		angleRender -= 360.0
	}
	// Calculate screenspace coordinates
	c.Distance = math.Sqrt(ebd.Dist2)
	xMid := math.Tan(angleRender*concepts.Deg2rad)*r.CameraToProjectionPlane + float64(r.ScreenWidth)*0.5
	depthScale := r.CameraToProjectionPlane / math.Cos(angleRender*concepts.Deg2rad)
	depthScale /= c.Distance
	xScale := depthScale * b.Size.Render[0]
	c.ScaleW = uint32(xScale)
	x1 := concepts.Max(int(xMid-xScale*0.5), xStart)
	x2 := concepts.Min(int(xMid+xScale*0.5), xEnd)
	if x1 == x2 || x2 < xStart || x1 >= xEnd {
		return
	}

	c.ProjectedTop = (b.Pos.Render[2] + b.Size.Render[1]*0.5 - c.CameraZ) * depthScale
	c.ProjectedBottom = c.ProjectedTop - b.Size.Render[1]*depthScale
	screenTop := c.ScreenHeight/2 - int(math.Floor(c.ProjectedTop))
	screenBottom := c.ScreenHeight/2 - int(math.Floor(c.ProjectedBottom))
	c.ScaleH = uint32(screenBottom - screenTop)
	c.ClippedTop = concepts.Clamp(screenTop, c.EdgeTop, c.EdgeBottom)
	c.ClippedBottom = concepts.Clamp(screenBottom, c.EdgeTop, c.EdgeBottom)

	if c.Pick &&
		c.ScreenX >= x1 && c.ScreenX < x2 &&
		c.ScreenY >= c.ClippedTop && c.ScreenY <= c.ClippedBottom {

		c.PickedSelection = append(c.PickedSelection, selection.SelectableFromBody(b))
		return
	}

	if lit := materials.GetLit(r.ECS, b.Entity); lit != nil {
		ls := &c.LightSampler
		ls.Sector = b.Sector()
		ls.Type = state.LightSamplerBody
		ls.MapIndex = c.WorldToLightmapAddress(ls.Sector, b.Pos.Render, uint16(ls.Type))
		ls.Segment = nil
		ls.InputBody = b.Entity
		ls.Get()
		c.Light[0] = ls.Output[0]
		c.Light[1] = ls.Output[1]
		c.Light[2] = ls.Output[2]
		c.Light[3] = 1
		// result = Surface * Diffuse * (Ambient + Lightmap)
		c.Light.To3D().AddSelf(&lit.Ambient)
		c.Light.Mul4Self(&lit.Diffuse)
	} else {
		c.Light[0] = 1
		c.Light[1] = 1
		c.Light[2] = 1
		c.Light[3] = 1
	}
	if alive := behaviors.GetAlive(r.ECS, b.Entity); alive != nil {
		alive.Tint(&c.Light)
	}

	vStart := float64(c.ScreenHeight/2) - c.ProjectedTop
	c.Light.MulSelf(ebd.Visible.Opacity)
	c.MaterialSampler.Initialize(b.Entity, nil)
	for c.ScreenX = x1; c.ScreenX < x2; c.ScreenX++ {
		if c.ScreenX < xStart || c.ScreenX >= xEnd {
			continue
		}
		c.MaterialSampler.NU = 0.5 + (float64(c.ScreenX)-xMid)/xScale

		for y := c.ClippedTop; y < c.ClippedBottom; y++ {
			screenIndex := (y*r.ScreenWidth + c.ScreenX)
			if c.Distance >= r.ZBuffer[screenIndex] {
				continue
			}
			c.NV = (float64(y) - vStart) / (c.ProjectedTop - c.ProjectedBottom)
			c.MaterialSampler.U = c.NU
			c.MaterialSampler.V = c.NV
			c.SampleMaterial(nil)
			c.MaterialSampler.Output.Mul4Self(&c.Light)
			r.ApplySample(&c.MaterialSampler.Output, screenIndex, c.Distance)
		}
	}
}

func (r *Renderer) renderBodyPixel(ebd *state.EntityWithDist2, c *state.Column, sx, ex int) {
	b := ebd.Body
	lit := materials.GetLit(r.ECS, b.Entity)
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
	c.Light.From(&lit.Diffuse)
	c.Light.To3D().AddSelf(&lit.Ambient)
	r.ApplySample(&c.Light, screenIndex, dist)
}
