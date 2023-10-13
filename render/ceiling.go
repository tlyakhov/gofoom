package render

import (
	"math"

	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/render/state"
)

// Ceiling renders the ceiling portion of a slice.
func Ceiling(s *state.Slice) {
	mat := s.PhysicalSector.CeilMaterial

	// Because of our sloped ceilings, we can't use simple linear interpolation to calculate the distance
	// or world position of the ceiling sample, we have to do a ray-plane intersection.
	// Thankfully, the only expensive operation is a square root to get the distance.
	planeRayDelta := s.PhysicalSector.Segments[0].P.Sub(&s.Ray.Start).To3D(&concepts.Vector3{})
	planeRayDelta[2] = s.PhysicalSector.TopZ - s.CameraZ
	rayDir := concepts.Vector3{s.AngleCos * s.ViewFix[s.X], s.AngleSin * s.ViewFix[s.X], 0}

	for s.Y = s.YStart; s.Y < s.ClippedStart; s.Y++ {
		rayDir[2] = float64(s.ScreenHeight/2 - 1 - s.Y)
		denom := s.PhysicalSector.CeilNormal.Dot(&rayDir)

		if math.Abs(denom) == 0 {
			continue
		}

		t := planeRayDelta.Dot(&s.PhysicalSector.CeilNormal) / denom
		world := (&concepts.Vector3{s.Ray.Start[0] + rayDir[0]*t, s.Ray.Start[1] + rayDir[1]*t, s.CameraZ + rayDir[2]*t})
		distToCeil := world.Length()
		scaler := s.PhysicalSector.CeilScale / distToCeil
		screenIndex := uint32(s.X + s.Y*s.ScreenWidth)

		if distToCeil >= s.ZBuffer[screenIndex] {
			continue
		}

		tx := world[0] / s.PhysicalSector.CeilScale
		tx -= math.Floor(tx)
		ty := world[1] / s.PhysicalSector.CeilScale
		ty -= math.Floor(ty)
		tx = math.Abs(tx)
		ty = math.Abs(ty)

		if mat != nil {
			s.Write(screenIndex, s.SampleMaterial(mat, tx, ty, s.Light(world, 0, 0), scaler))
		}
		s.ZBuffer[screenIndex] = distToCeil
	}
}
