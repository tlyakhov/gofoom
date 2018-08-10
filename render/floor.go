package render

import (
	"math"

	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/render/material"
	"github.com/tlyakhov/gofoom/render/state"
)

func Floor(s *state.Slice) {
	mat := material.For(s.PhysicalSector.FloorMaterial, s)

	world := &concepts.Vector3{0, 0, s.PhysicalSector.BottomZ}

	for s.Y = s.ClippedEnd; s.Y < s.YEnd; s.Y++ {
		if s.Y-s.ScreenHeight/2 == 0 {
			continue
		}

		distToFloor := (-s.PhysicalSector.BottomZ + s.CameraZ) * s.ViewFix[s.X] / float64(s.Y-s.ScreenHeight/2)
		scaler := s.PhysicalSector.FloorScale / distToFloor
		screenIndex := uint32(s.X + s.Y*s.ScreenWidth)

		if distToFloor >= s.ZBuffer[screenIndex] {
			continue
		}

		world.X = s.Map.Player.Physical().Pos.X + s.AngleCos*distToFloor
		world.Y = s.Map.Player.Physical().Pos.Y + s.AngleSin*distToFloor

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
			s.Write(screenIndex, mat.Sample(tx, ty, s.Light(world, &state.FloorNormal, 0, 0), scaler))
		}
		s.ZBuffer[screenIndex] = distToFloor
	}
}
