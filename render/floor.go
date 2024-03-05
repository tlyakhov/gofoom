package render

import (
	"fmt"

	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/render/state"
)

func FloorPick(s *state.Column) {
	if s.Y >= s.ClippedEnd && s.Y < s.YEnd {
		s.PickedElements = append(s.PickedElements, state.PickedElement{Type: state.PickFloor, Element: s.Sector.Ref()})
	}
}

// Floor renders the floor portion of a slice.
func Floor(c *state.Column) {
	mat := c.Sector.FloorSurface.Material
	extras := c.Sector.FloorSurface.ExtraStages
	transform := c.Sector.FloorSurface.Transform

	// Because of our sloped floors, we can't use simple linear interpolation to calculate the distance
	// or world position of the floor sample, we have to do a ray-plane intersection.
	// Thankfully, the only expensive operation is a square root to get the distance.
	planeRayDelta := &concepts.Vector3{c.Sector.Segments[0].P[0] - c.Ray.Start[0], c.Sector.Segments[0].P[1] - c.Ray.Start[1], c.Sector.BottomZ.Render - c.CameraZ}
	rayDir := &concepts.Vector3{c.AngleCos * c.ViewFix[c.X], c.AngleSin * c.ViewFix[c.X], 0}
	for c.Y = c.ClippedEnd; c.Y < c.YEnd; c.Y++ {
		rayDir[2] = float64(c.ScreenHeight/2 - c.Y)
		denom := c.Sector.FloorNormal.Dot(rayDir)

		if denom == 0 {
			continue
		}

		t := planeRayDelta.Dot(&c.Sector.FloorNormal) / denom
		if t <= 0 {
			//s.Write(uint32(s.X+s.Y*s.ScreenWidth), 255)
			dbg := fmt.Sprintf("%v floor t <= 0", c.Sector.Entity)
			c.DebugNotices.Push(dbg)
			continue
		}
		world := &concepts.Vector3{rayDir[0] * t, rayDir[1] * t, rayDir[2] * t}
		distToFloor := world.Length()
		dist2 := world.To2D().Length2()
		screenIndex := uint32(c.X + c.Y*c.ScreenWidth)

		if distToFloor > c.ZBuffer[screenIndex] || dist2 > c.Distance*c.Distance {
			continue
		}

		world[0] += c.Ray.Start[0]
		world[1] += c.Ray.Start[1]
		world[2] += c.CameraZ
		scaler := 64.0 / distToFloor
		tx := world[0] / 64.0
		ty := world[1] / 64.0

		if !mat.Nil() {
			tx, ty = transform[0]*tx+transform[2]*ty+transform[4], transform[1]*tx+transform[3]*ty+transform[5]
			c.SampleShader(mat, extras, tx, ty, scaler)
			c.SampleLight(&c.MaterialSampler.Output, mat, world, 0, 0, distToFloor)
		}
		//concepts.AsmVector4AddPreMulColorSelf((*[4]float64)(&s.FrameBuffer[screenIndex]), (*[4]float64)(&s.Material))
		c.FrameBuffer[screenIndex].AddPreMulColorSelf(&c.MaterialSampler.Output)
		c.ZBuffer[screenIndex] = distToFloor
	}
}
