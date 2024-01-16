package render

import (
	"math"
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
	billboard := false
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
		billboard = sheet.Billboard
	}

	var scaler float64
	if billboard {
		angleRender := angleFromPlayer - r.PlayerBody.Angle.Render
		if angleRender < -180.0 {
			angleRender += 360.0
		}
		if angleRender > 180.0 {
			angleRender -= 360.0
		}
		c.Distance = c.Ray.Start.Dist(b.Pos.Render.To2D())
		x := (angleRender + r.FOV*0.5) * float64(r.ScreenWidth) / r.FOV

		vfixindex := concepts.IntClamp(int(x), 0, r.ScreenWidth-1)
		scaler = r.ViewFix[vfixindex] / c.Distance
		scale := scaler * b.Size.Render[0]
		x1 := (x - scale*0.5)
		x2 := x1 + scale
		c.U = 0.5 + (float64(c.X)-x)/scale
		if x1 > float64(c.X) || x2 < float64(c.X) {
			return
		}
	} else {
		isect := new(concepts.Vector2)
		ray := new(concepts.Vector2)
		ray.From(&c.Ray.End).SubSelf(&c.Ray.Start)
		sx1 := &concepts.Vector2{
			-math.Sin(b.Angle.Render*concepts.Deg2rad) * b.Size.Render[0] * 0.5,
			math.Cos(b.Angle.Render*concepts.Deg2rad) * b.Size.Render[0] * 0.5}
		sx2 := &concepts.Vector2{b.Pos.Render[0] + sx1[0], b.Pos.Render[1] + sx1[1]}
		sx1[0] = b.Pos.Render[0] - sx1[0]
		sx1[1] = b.Pos.Render[1] - sx1[1]

		/*// Wall is facing away from us
		if ray.Dot(&segment.Normal) > 0 {
			continue
		}*/

		// Ray intersects?
		if ok := concepts.IntersectSegments(sx1, sx2, &c.Ray.Start, &c.Ray.End, isect); !ok {
			return
		}

		delta := concepts.Vector2{math.Abs(isect[0] - c.Ray.Start[0]), math.Abs(isect[1] - c.Ray.Start[1])}
		if delta[1] > delta[0] {
			c.Distance = math.Abs(delta[1] / c.AngleSin)
		} else {
			c.Distance = math.Abs(delta[0] / c.AngleCos)
		}
		isect.To3D(&c.Intersection)
		c.U = isect.Dist(sx1) / b.Size.Render[0]
	}

	img := materials.ImageFromDb(refMaterial)
	if img == nil {
		return
	}

	c.ProjHeightTop = c.ProjectZ(b.Pos.Render[2] + b.Size.Render[1] - c.CameraZ)
	c.ProjHeightBottom = c.ProjectZ(b.Pos.Render[2] - c.CameraZ)
	c.ScreenStart = c.ScreenHeight/2 - int(c.ProjHeightTop)
	c.ScreenEnd = c.ScreenHeight/2 - int(c.ProjHeightBottom)
	c.ClippedStart = concepts.IntClamp(c.ScreenStart, c.YStart, c.YEnd)
	c.ClippedEnd = concepts.IntClamp(c.ScreenEnd, c.YStart, c.YEnd)

	if c.Pick && c.Y >= c.ClippedStart && c.Y <= c.ClippedEnd {
		c.PickedElements = append(c.PickedElements, state.PickedElement{Type: state.PickBody, Element: ref})
		return
	}

	if lit := materials.LitFromDb(ref); lit != nil {
		le := &c.LightElements[0]
		le.Q.From(&b.Pos.Render)
		le.Q[2] += b.Size.Render[1] * 0.5
		le.Lightmap = nil
		le.LightmapAge = nil
		le.MapIndex = 0
		le.Segment = nil
		le.Type = state.LightElementBody
		le.InputBody = ref
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
		sample := c.SampleShader(refMaterial, nil, c.U, v, scaler)
		sample.Mul4Self(&c.Light)
		if sample[3] > 0 {
			c.FrameBuffer[screenIndex].AddPreMulColorSelf(sample)
			r.ZBuffer[screenIndex] = c.Distance
		}
	}
}
