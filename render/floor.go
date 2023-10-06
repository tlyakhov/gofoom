package render

import (
	"math"

	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/render/material"
	"tlyakhov/gofoom/render/state"
)

// Floor renders the floor portion of a slice.
func Floor(s *state.Slice) {
	mat := material.For(s.PhysicalSector.FloorMaterial, s)

	// Because of our sloped floors, we can't use simple linear interpolation to calculate the distance
	// or world position of the floor sample, we have to do a ray-plane intersection.
	// Thankfully, the only expensive operation is a square root to get the distance.
	planeRayDelta := s.PhysicalSector.Segments[0].P.Sub(s.Ray.Start).To3D()
	planeRayDelta.Z = s.PhysicalSector.BottomZ - s.CameraZ
	rayDir := concepts.Vector3{X: s.AngleCos * s.ViewFix[s.X], Y: s.AngleSin * s.ViewFix[s.X], Z: 0}

	for s.Y = s.ClippedEnd; s.Y < s.YEnd; s.Y++ {
		rayDir.Z = float64(s.ScreenHeight/2 - s.Y)
		denom := s.PhysicalSector.FloorNormal.Dot(rayDir)

		if math.Abs(denom) == 0 {
			continue
		}

		t := planeRayDelta.Dot(s.PhysicalSector.FloorNormal) / denom
		world := concepts.Vector3{X: s.Ray.Start.X, Y: s.Ray.Start.Y, Z: s.CameraZ}.Add(rayDir.Mul(t))
		distToFloor := world.Length()
		scaler := s.PhysicalSector.FloorScale / distToFloor
		screenIndex := uint32(s.X + s.Y*s.ScreenWidth)

		if distToFloor >= s.ZBuffer[screenIndex] {
			continue
		}

		tx := world.X / s.PhysicalSector.FloorScale
		tx -= math.Floor(tx)
		ty := world.Y / s.PhysicalSector.FloorScale
		ty -= math.Floor(ty)
		if tx < 0 {
			tx += 1.0
		}
		if ty < 0 {
			ty += 1.0
		}

		if mat != nil {
			s.Write(screenIndex, mat.Sample(tx, ty, s.Light(world, s.PhysicalSector.FloorNormal, 0, 0), scaler))
		}
		s.ZBuffer[screenIndex] = distToFloor
	}
}
