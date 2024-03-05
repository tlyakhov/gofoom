package render

import (
	"math"
	"tlyakhov/gofoom/archetypes"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/render/state"
)

func (r *Renderer) RenderBody(ref *concepts.EntityRef, c *state.Column) {
	// TODO: We should probably not do all of this for every column. Can we
	// cache any of these things as we cast rays?
	b := core.BodyFromDb(ref)
	if b == nil || !b.IsActive() {
		return
	}

	refMaterial := ref
	angleFromPlayer := r.PlayerBody.Angle2DTo(&b.Pos.Render)

	sheet := materials.SpriteSheetFromDb(ref)
	if sheet != nil && sheet.IsActive() {
		angleSprite := 360 - angleFromPlayer + b.Angle.Render
		if len(sheet.Sprites) == 0 {
			return
		}
		// Pick the sprite to use
		// Too slow?
		bestDelta := 1000.0
		var sprite *materials.Sprite
		for i, s := range sheet.Sprites {
			if s == nil {
				continue
			}
			angleDelta := float64(s.Angle) - angleSprite
			for angleDelta < -180 {
				angleDelta += 360
			}
			for angleDelta > 180 {
				angleDelta -= 360
			}
			angleDelta = math.Abs(angleDelta)
			if sprite == nil || angleDelta < bestDelta {
				bestDelta = angleDelta
				sprite = sheet.Sprites[i]
			}
		}

		if sprite == nil {
			return
		}
		refMaterial = sprite.Image
	}

	if !archetypes.EntityRefIsMaterial(refMaterial) {
		return
	}

	angleRender := angleFromPlayer - r.PlayerBody.Angle.Render
	for angleRender < -180.0 {
		angleRender += 360.0
	}
	for angleRender > 180.0 {
		angleRender -= 360.0
	}
	c.Distance = c.Ray.Start.Dist(b.Pos.Render.To2D())
	x := (angleRender + r.FOV*0.5) * float64(r.ScreenWidth) / r.FOV

	vfixindex := concepts.IntClamp(int(x), 0, r.ScreenWidth-1)
	depthScale := r.ViewFix[vfixindex] / c.Distance
	xScale := depthScale * b.Size.Render[0]
	x1 := (x - xScale*0.5)
	x2 := x1 + xScale
	c.U = 0.5 + (float64(c.X)-x)/xScale
	if x1 > float64(c.X) || x2 < float64(c.X) {
		return
	}

	c.ProjHeightTop = (b.Pos.Render[2] + b.Size.Render[1] - c.CameraZ) * depthScale
	c.ProjHeightBottom = (b.Pos.Render[2] - c.CameraZ) * depthScale
	c.ScreenStart = c.ScreenHeight/2 - int(c.ProjHeightTop)
	c.ScreenEnd = c.ScreenHeight/2 - int(c.ProjHeightBottom)
	c.ClippedStart = concepts.IntClamp(c.ScreenStart, c.YStart, c.YEnd)
	c.ClippedEnd = concepts.IntClamp(c.ScreenEnd, c.YStart, c.YEnd)

	if c.Pick && c.Y >= c.ClippedStart && c.Y <= c.ClippedEnd {
		c.PickedElements = append(c.PickedElements, state.PickedElement{Type: state.PickBody, Element: ref})
		return
	}

	if lit := materials.LitFromDb(ref); lit != nil {
		le := &c.LightElement
		le.Q.From(&b.Pos.Render)
		le.Q[2] += b.Size.Render[1] * 0.5
		le.MapIndex = b.Sector().WorldToLightmapAddress(&le.Q, 0)
		le.Segment = nil
		le.Type = state.LightElementBody
		le.InputBody = ref
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

	for y := c.ClippedStart; y < c.ClippedEnd; y++ {
		screenIndex := (y*r.ScreenWidth + c.X)
		if c.Distance >= r.ZBuffer[screenIndex] {
			continue
		}
		v := float64(y-c.ScreenStart) / float64(c.ScreenEnd-c.ScreenStart)
		sample := c.SampleShader(refMaterial, nil, c.U, v, depthScale)
		sample.Mul4Self(&c.Light)
		c.ApplySample(sample, screenIndex, c.Distance)
	}
}
