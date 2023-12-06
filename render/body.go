package render

import (
	"math"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/render/state"
)

func (r *Renderer) RenderBody(ref *concepts.EntityRef, slice *state.Slice) {
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
	y2 := (float64(r.ScreenHeight)*0.5 - (b.Pos.Render[2]-slice.CameraZ)*scaler)
	y1 := y2 - b.Height*scaler

	bodyWidth := 32.0
	scale := scaler * bodyWidth
	x1 := (x - scale*0.5)
	x2 := x1 + scale

	if x1 > float64(slice.X) || x2 < float64(slice.X) {
		return
	}
	clippedY1 := int(y1)
	if clippedY1 < 0 {
		clippedY1 = 0
	}
	clippedY2 := int(r.ScreenHeight - 1)
	if clippedY2 > int(y2) {
		clippedY2 = int(y2)
	}

	u := 0.5 + (float64(slice.X)-x)/scale

	le := &slice.LightElements[0]
	le.Normal[0] = math.Cos(b.Angle * concepts.Deg2rad)
	le.Normal[1] = math.Sin(b.Angle * concepts.Deg2rad)
	le.Normal[2] = 0
	le.Lightmap = make([]concepts.Vector3, 1)
	le.LightmapAge = make([]int, 1)
	le.MapIndex = 0

	light := le.Get(true)

	for y := clippedY1; y < clippedY2; y++ {
		screenIndex := (y*r.ScreenWidth + slice.X)
		if d >= r.ZBuffer[screenIndex] {
			continue
		}
		v := (float64(y) - y1) / (y2 - y1)
		c := slice.SampleMaterial(ref, u, v, light, scaler)
		if c[3] > 0 {
			slice.Write(uint32(screenIndex), c)
			r.ZBuffer[screenIndex] = d
		}
	}
}
