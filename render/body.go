package render

import (
	"math"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/render/state"
)

func (r *Renderer) RenderBody(ref *concepts.EntityRef, s *state.Column) {
	b := core.BodyFromDb(ref)
	sheet := materials.SpriteSheetFromDb(ref)
	if b == nil || sheet == nil {
		return
	}
	angleFromPlayer := r.PlayerBody.Angle2DTo(&b.Pos.Render)
	angleRender := angleFromPlayer - r.PlayerBody.Angle
	angleSprite := 360 - angleFromPlayer + b.Angle

	if angleRender < -180.0 {
		angleRender += 360.0
	}
	if angleRender > 180.0 {
		angleRender -= 360.0
	}

	d := r.PlayerBody.Pos.Render.Dist(&b.Pos.Render)
	x := (angleRender + r.FOV*0.5) * float64(r.ScreenWidth) / r.FOV

	vfixindex := concepts.IntClamp(int(x), 0, r.ScreenWidth-1)
	scaler := r.ViewFix[vfixindex] / d
	y2 := (float64(r.ScreenHeight)*0.5 - (b.Pos.Render[2]-s.CameraZ)*scaler)
	y1 := y2 - b.Size.Render[2]*scaler

	if len(sheet.Sprites) == 0 {
		return
	}

	// Pick the sprite to use
	bestDelta := 1000.0
	var sprite *materials.Sprite
	for i, s := range sheet.Sprites {
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
			sprite = &sheet.Sprites[i]
		}
	}
	/*dbg := fmt.Sprintf("%v - sprite:%.1f, delta:%.1f s:%v", ref.String(), angleSprite, bestDelta, sprite.Angle)
	s.DebugNotices.Push(dbg)*/

	refMaterial := sprite.Image
	img := materials.ImageFromDb(refMaterial)
	if img == nil {
		return
	}
	scale := scaler * b.Size.Render[0]
	x1 := (x - scale*0.5)
	x2 := x1 + scale

	if x1 > float64(s.X) || x2 < float64(s.X) {
		return
	}
	clippedY1 := concepts.Max(int(y1), 0)
	clippedY2 := concepts.Min(int(y2), int(r.ScreenHeight-1))

	if s.Pick && s.Y >= clippedY1 && s.Y <= clippedY2 {
		s.PickedElements = append(s.PickedElements, state.PickedElement{Type: state.PickBody, Element: ref})
		return
	}

	u := 0.5 + (float64(s.X)-x)/scale

	if lit := materials.LitFromDb(ref); lit != nil {
		le := &s.LightElements[0]
		le.Q.From(&b.Pos.Render)
		le.Q[2] += b.Size.Render[2] * 0.5
		le.Lightmap = nil
		le.LightmapAge = nil
		le.MapIndex = 0
		le.Segment = nil
		le.Type = state.LightElementBody
		le.InputBody = ref
		le.Get()
		s.Light.To3D().From(&le.Output)
		s.Light[3] = 1
		// result = Surface * Diffuse * (Ambient + Lightmap)
		s.Light.To3D().AddSelf(&lit.Ambient)
		s.Light.Mul4Self(&lit.Diffuse)
	} else {
		s.Light[0] = 1
		s.Light[1] = 1
		s.Light[2] = 1
		s.Light[3] = 1
	}

	for y := clippedY1; y < clippedY2; y++ {
		screenIndex := (y*r.ScreenWidth + s.X)
		if d >= r.ZBuffer[screenIndex] {
			continue
		}
		v := (float64(y) - y1) / (y2 - y1)
		c := s.SampleMaterial(refMaterial, u, v, scaler)
		c.Mul4Self(&s.Light)
		if c[3] > 0 {
			s.FrameBuffer[screenIndex].AddPreMulColorSelf(c)
			r.ZBuffer[screenIndex] = d
		}
	}
}
