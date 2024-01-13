package render

import (
	"fmt"
	"math"

	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/render/state"
)

func CeilingPick(s *state.Column) {
	if s.Y >= s.YStart && s.Y < s.ClippedStart {
		s.PickedElements = append(s.PickedElements, state.PickedElement{Type: state.PickCeiling, Element: s.Sector.Ref()})
	}
}

// Ceiling renders the ceiling portion of a slice.
func Ceiling(s *state.Column) {
	mat := s.Sector.CeilSurface.Material
	extras := s.Sector.CeilSurface.ExtraStages
	transform := s.Sector.CeilSurface.Transform

	// Because of our sloped ceilings, we can't use simple linear interpolation to calculate the distance
	// or world position of the ceiling sample, we have to do a ray-plane intersection.
	// Thankfully, the only expensive operation is a square root to get the distance.
	planeRayDelta := &concepts.Vector3{s.Sector.Segments[0].P[0] - s.Ray.Start[0], s.Sector.Segments[0].P[1] - s.Ray.Start[1], s.Sector.TopZ.Render - s.CameraZ}
	rayDir := concepts.Vector3{s.AngleCos * s.ViewFix[s.X], s.AngleSin * s.ViewFix[s.X], 0}
	for s.Y = s.YStart; s.Y < s.ClippedStart; s.Y++ {
		rayDir[2] = float64(s.ScreenHeight/2 - 1 - s.Y)
		denom := s.Sector.CeilNormal.Dot(&rayDir)

		if math.Abs(denom) == 0 {
			continue
		}

		t := planeRayDelta.Dot(&s.Sector.CeilNormal) / denom
		if t <= 0 {
			//s.Write(uint32(s.X+s.Y*s.ScreenWidth), 255)
			dbg := fmt.Sprintf("%v ceiling t <= 0", s.Sector.Entity)
			s.DebugNotices.Push(dbg)
			continue
		}
		world := &concepts.Vector3{rayDir[0] * t, rayDir[1] * t, rayDir[2] * t}
		distToCeil := world.Length()
		world[0] += s.Ray.Start[0]
		world[1] += s.Ray.Start[1]
		world[2] += s.CameraZ
		scaler := 64.0 / distToCeil
		screenIndex := uint32(s.X + s.Y*s.ScreenWidth)

		if distToCeil >= s.ZBuffer[screenIndex] {
			continue
		}

		tx := world[0] / 64.0
		ty := world[1] / 64.0

		if !mat.Nil() {
			tx, ty = transform[0]*tx+transform[2]*ty+transform[4], transform[1]*tx+transform[3]*ty+transform[5]
			s.SampleShader(mat, extras, tx, ty, scaler)
			s.SampleLight(&s.MaterialColor, mat, world, 0, 0, distToCeil)
		}
		//concepts.AsmVector4AddPreMulColorSelf((*[4]float64)(&s.FrameBuffer[screenIndex]), (*[4]float64)(&s.Material))
		s.FrameBuffer[screenIndex].AddPreMulColorSelf(&s.MaterialColor)
		s.ZBuffer[screenIndex] = distToCeil
	}
}
