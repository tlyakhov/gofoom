package render

import (
	"math"

	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/render/state"
)

// Floor renders the floor portion of a slice.
func Floor(s *state.Slice) {
	mat := s.PhysicalSector.FloorMaterial

	// Because of our sloped floors, we can't use simple linear interpolation to calculate the distance
	// or world position of the floor sample, we have to do a ray-plane intersection.
	// Thankfully, the only expensive operation is a square root to get the distance.
	planeRayDelta := s.PhysicalSector.Segments[0].P.Sub(&s.Ray.Start).To3D(&concepts.Vector3{})
	planeRayDelta[2] = s.PhysicalSector.BottomZ - s.CameraZ
	rayDir := &concepts.Vector3{s.AngleCos * s.ViewFix[s.X], s.AngleSin * s.ViewFix[s.X], 0}

	for s.Y = s.ClippedEnd; s.Y < s.YEnd; s.Y++ {
		rayDir[2] = float64(s.ScreenHeight/2 - s.Y)
		denom := s.PhysicalSector.FloorNormal.Dot(rayDir)

		if math.Abs(denom) == 0 {
			continue
		}

		t := planeRayDelta.Dot(&s.PhysicalSector.FloorNormal) / denom
		world := (&concepts.Vector3{s.Ray.Start[0] + rayDir[0]*t, s.Ray.Start[1] + rayDir[1]*t, s.CameraZ + rayDir[2]*t})
		distToFloor := world.Length()
		scaler := s.PhysicalSector.FloorScale / distToFloor
		screenIndex := uint32(s.X + s.Y*s.ScreenWidth)

		if distToFloor >= s.ZBuffer[screenIndex] {
			continue
		}

		tx := world[0] / s.PhysicalSector.FloorScale
		tx -= math.Floor(tx)
		ty := world[1] / s.PhysicalSector.FloorScale
		ty -= math.Floor(ty)
		if tx < 0 {
			tx += 1.0
		}
		if ty < 0 {
			ty += 1.0
		}

		if mat != nil {
			s.Write(screenIndex, s.SampleMaterial(mat, tx, ty, s.Light(world, 0, 0), scaler))
		}
		s.ZBuffer[screenIndex] = distToFloor
	}
}
