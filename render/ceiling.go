package render

import (
	"math"

	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/render/state"
)

func CeilingPick(s *state.Slice) {
	if s.Y >= s.YStart && s.Y < s.ClippedStart {
		s.PickedElements = append(s.PickedElements, state.PickedElement{Type: "ceiling", ISerializable: s.PhysicalSector})
	}
}

// Ceiling renders the ceiling portion of a slice.
func Ceiling(s *state.Slice) {
	mat := s.PhysicalSector.CeilMaterial

	// Because of our sloped ceilings, we can't use simple linear interpolation to calculate the distance
	// or world position of the ceiling sample, we have to do a ray-plane intersection.
	// Thankfully, the only expensive operation is a square root to get the distance.
	planeRayDelta := &concepts.Vector3{s.PhysicalSector.Segments[0].P[0] - s.Ray.Start[0], s.PhysicalSector.Segments[0].P[1] - s.Ray.Start[1], s.PhysicalSector.TopZ - s.CameraZ}
	rayDir := concepts.Vector3{s.AngleCos * s.ViewFix[s.X], s.AngleSin * s.ViewFix[s.X], 0}
	light := concepts.Vector3{}
	for s.Y = s.YStart; s.Y < s.ClippedStart; s.Y++ {
		rayDir[2] = float64(s.ScreenHeight/2 - 1 - s.Y)
		denom := s.PhysicalSector.CeilNormal.Dot(&rayDir)

		if math.Abs(denom) == 0 {
			continue
		}

		t := planeRayDelta.Dot(&s.PhysicalSector.CeilNormal) / denom
		if t <= 0 {
			//s.Write(uint32(s.X+s.Y*s.ScreenWidth), 255)
			//log.Printf("ceil t<0\n")
			continue
		}
		world := &concepts.Vector3{rayDir[0] * t, rayDir[1] * t, rayDir[2] * t}
		distToCeil := world.Length()
		world[0] += s.Ray.Start[0]
		world[1] += s.Ray.Start[1]
		world[2] += s.CameraZ
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
			s.Write(screenIndex, s.SampleMaterial(mat, tx, ty, s.Light(&light, world, 0, 0, distToCeil), scaler))
		}
		s.ZBuffer[screenIndex] = distToCeil
	}
}
