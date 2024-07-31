// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package render

import (
	"math"
	"tlyakhov/gofoom/archetypes"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/render/state"
)

func (r *Renderer) renderBody(b *core.Body, c *state.Column) {
	if !archetypes.EntityIsMaterial(r.DB, b.Entity) {
		return
	}
	// TODO: We should probably not do all of this for every column. Can we
	// cache any of these things as we cast rays?

	// Calculate angles for picking the right sprite, and also relative to camera
	angleFromPlayer := r.PlayerBody.Angle2DTo(b.Pos.Render)
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
	c.Distance = c.Ray.Start.Dist(b.Pos.Render.To2D())
	x := math.Tan(angleRender*concepts.Deg2rad)*r.CameraToProjectionPlane + float64(r.ScreenWidth)*0.5
	depthScale := r.CameraToProjectionPlane / math.Cos(angleRender*concepts.Deg2rad)
	depthScale /= c.Distance
	xScale := depthScale * b.Size.Render[0]
	x1 := (x - xScale*0.5)
	x2 := x1 + xScale
	c.U = 0.5 + (float64(c.ScreenX)-x)/xScale
	if x1 > float64(c.ScreenX) || x2 < float64(c.ScreenX) {
		return
	}

	c.ProjectedTop = (b.Pos.Render[2] + b.Size.Render[1]*0.5 - c.CameraZ) * depthScale
	c.ProjectedBottom = c.ProjectedTop - b.Size.Render[1]*depthScale
	screenTop := c.ScreenHeight/2 - int(c.ProjectedTop)
	screenBottom := c.ScreenHeight/2 - int(c.ProjectedBottom)
	c.ClippedTop = concepts.Clamp(screenTop, c.EdgeTop, c.EdgeBottom)
	c.ClippedBottom = concepts.Clamp(screenBottom, c.EdgeTop, c.EdgeBottom)

	if c.Pick && c.ScreenY >= c.ClippedTop && c.ScreenY <= c.ClippedBottom {
		c.PickedSelection = append(c.PickedSelection, core.SelectableFromBody(b))
		return
	}

	if lit := materials.LitFromDb(r.DB, b.Entity); lit != nil {
		le := &c.LightElement
		le.Q.From(b.Pos.Render)
		le.MapIndex = b.Sector().WorldToLightmapAddress(&le.Q, 0)
		le.Segment = nil
		le.Type = state.LightElementBody
		le.InputBody = b.Entity
		le.Sector = b.Sector()
		le.Get()
		c.Light.To3D().From(&le.Output)
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
	vStart := float64(c.ScreenHeight/2) - c.ProjectedTop
	seen := false
	for y := c.ClippedTop; y < c.ClippedBottom; y++ {
		screenIndex := (y*r.ScreenWidth + c.ScreenX)
		if c.Distance >= r.ZBuffer[screenIndex] {
			continue
		}
		v := (float64(y) - vStart) / (c.ProjectedTop - c.ProjectedBottom)
		sample := c.SampleShader(b.Entity, nil, c.U, v, uint32(xScale), uint32(screenBottom-screenTop))
		sample.Mul4Self(&c.Light)
		r.ApplySample(sample, screenIndex, c.Distance)
		seen = true
	}
	if seen {
		c.BodiesSeen[b.Entity] = b
	}
}
