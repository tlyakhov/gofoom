package render

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/render/state"
)

func (r *Renderer) RenderBody(ref *concepts.EntityRef, s *state.Slice) {
	b := core.BodyFromDb(ref)
	img := materials.ImageFromDb(ref)
	if b == nil || img == nil {
		return
	}
	renderAngle := r.PlayerBody.Angle2DTo(&b.Pos.Render)
	bodyAngle := renderAngle - r.PlayerBody.Angle

	if bodyAngle < -r.FOV/2 {
		bodyAngle += 360.0
	}
	if bodyAngle > r.FOV/2 {
		bodyAngle -= 360.0
	}

	d := r.PlayerBody.Pos.Render.Dist(&b.Pos.Render)
	x := (bodyAngle + r.FOV*0.5) * float64(r.ScreenWidth) / r.FOV
	vfixindex := concepts.IntClamp(int(x), 0, r.ScreenWidth-1)
	scaler := r.ViewFix[vfixindex] / d
	y2 := (float64(r.ScreenHeight)*0.5 - (b.Pos.Render[2]-s.CameraZ)*scaler)
	y1 := y2 - b.Height*scaler

	bodyWidth := float64(img.Width) * b.Height / float64(img.Height)
	scale := scaler * bodyWidth
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

	le := &s.LightElements[0]
	le.Q.From(&b.Pos.Render)
	le.Q[2] += b.Height * 0.5
	le.Lightmap = nil
	le.LightmapAge = nil
	le.MapIndex = 0
	le.Segment = nil
	le.Type = state.LightElementBody
	light := le.Get()

	for y := clippedY1; y < clippedY2; y++ {
		screenIndex := (y*r.ScreenWidth + s.X)
		if d >= r.ZBuffer[screenIndex] {
			continue
		}
		v := (float64(y) - y1) / (y2 - y1)
		c := s.SampleMaterial(ref, u, v, light, scaler)
		if c[3] > 0 {
			s.Write(uint32(screenIndex), c)
			r.ZBuffer[screenIndex] = d
		}
	}
}
