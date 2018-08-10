package render

import (
	"math"

	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/render/material"
	"github.com/tlyakhov/gofoom/render/state"
)

func Ceiling(s *state.Slice) {
	mat := material.For(s.PhysicalSector.CeilMaterial, s)

	world := concepts.Vector3{0, 0, s.PhysicalSector.TopZ}

	for s.Y = s.YStart; s.Y < s.ClippedStart; s.Y++ {
		if s.Y-s.ScreenHeight/2 == 0 {
			continue
		}

		distToCeil := (s.PhysicalSector.TopZ - s.CameraZ) * s.ViewFix[s.X] / float64(s.ScreenHeight/2-1-s.Y)
		scaler := s.PhysicalSector.CeilScale / distToCeil
		screenIndex := uint32(s.X + s.Y*s.ScreenWidth)

		if distToCeil >= s.ZBuffer[screenIndex] {
			continue
		}

		world.X = s.Map.Player.Physical().Pos.X + s.AngleCos*distToCeil
		world.Y = s.Map.Player.Physical().Pos.Y + s.AngleSin*distToCeil

		tx := world.X / s.PhysicalSector.CeilScale
		tx -= math.Floor(tx)
		ty := world.Y / s.PhysicalSector.CeilScale
		ty -= math.Floor(ty)
		tx = math.Abs(tx)
		ty = math.Abs(ty)

		if mat != nil {
			s.Write(screenIndex, mat.Sample(tx, ty, s.Light(world, state.CeilingNormal, 0, 0), scaler))
		}
		s.ZBuffer[screenIndex] = distToCeil
	}
}
