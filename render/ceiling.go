package render

import (
	"fmt"
	"math"

	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/render/state"
)

func CeilingPick(s *state.Slice) {
	if s.Y >= s.YStart && s.Y < s.ClippedStart {
		s.PickedElements = append(s.PickedElements, state.PickedElement{Type: state.PickCeiling, Element: s.Sector.Ref()})
	}
}

// Ceiling renders the ceiling portion of a slice.
func Ceiling(s *state.Slice) {
	mat := s.Sector.CeilMaterial

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
		scaler := s.Sector.CeilScale / distToCeil
		screenIndex := uint32(s.X + s.Y*s.ScreenWidth)

		if distToCeil >= s.ZBuffer[screenIndex] {
			continue
		}

		tx := world[0] / s.Sector.CeilScale
		ty := world[1] / s.Sector.CeilScale

		c := &concepts.Vector4{0.5, 0.5, 0.5, 1}
		if !mat.Nil() {
			*c = s.SampleMaterial(mat, tx, ty, scaler)
			s.SampleLight(c, mat, world, 0, 0, distToCeil)
		}
		s.FrameBuffer[screenIndex].AddPreMulColorSelf(c)
		s.ZBuffer[screenIndex] = distToCeil
	}
}
