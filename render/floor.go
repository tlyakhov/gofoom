package render

import (
	"fmt"
	"math"

	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/render/state"
)

func FloorPick(s *state.Slice) {
	if s.Y >= s.ClippedEnd && s.Y < s.YEnd {
		s.PickedElements = append(s.PickedElements, state.PickedElement{Type: state.PickFloor, Element: s.Sector.Ref()})
	}
}

// Floor renders the floor portion of a slice.
func Floor(s *state.Slice) {
	mat := s.Sector.FloorMaterial

	// Because of our sloped floors, we can't use simple linear interpolation to calculate the distance
	// or world position of the floor sample, we have to do a ray-plane intersection.
	// Thankfully, the only expensive operation is a square root to get the distance.
	planeRayDelta := &concepts.Vector3{s.Sector.Segments[0].P[0] - s.Ray.Start[0], s.Sector.Segments[0].P[1] - s.Ray.Start[1], s.Sector.BottomZ.Render - s.CameraZ}
	rayDir := &concepts.Vector3{s.AngleCos * s.ViewFix[s.X], s.AngleSin * s.ViewFix[s.X], 0}
	light := concepts.Vector3{}
	for s.Y = s.ClippedEnd; s.Y < s.YEnd; s.Y++ {
		rayDir[2] = float64(s.ScreenHeight/2 - s.Y)
		denom := s.Sector.FloorNormal.Dot(rayDir)

		if math.Abs(denom) == 0 {
			continue
		}

		t := planeRayDelta.Dot(&s.Sector.FloorNormal) / denom
		if t <= 0 {
			//s.Write(uint32(s.X+s.Y*s.ScreenWidth), 255)
			dbg := fmt.Sprintf("%v floor t <= 0", s.Sector.Entity)
			s.DebugNotices.Push(dbg)
			continue
		}
		world := &concepts.Vector3{rayDir[0] * t, rayDir[1] * t, rayDir[2] * t}
		distToFloor := world.Length()
		world[0] += s.Ray.Start[0]
		world[1] += s.Ray.Start[1]
		world[2] += s.CameraZ
		scaler := s.Sector.FloorScale / distToFloor
		screenIndex := uint32(s.X + s.Y*s.ScreenWidth)

		if distToFloor >= s.ZBuffer[screenIndex] {
			continue
		}

		tx := world[0] / s.Sector.FloorScale
		ty := world[1] / s.Sector.FloorScale

		if !mat.Nil() {
			s.Write(screenIndex, s.SampleMaterial(mat, tx, ty, s.Light(&light, world, 0, 0, distToFloor), scaler))
		}
		s.ZBuffer[screenIndex] = distToFloor
	}
}
