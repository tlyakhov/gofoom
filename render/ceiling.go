package render

import (
	"math"

	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/render/material"
	"github.com/tlyakhov/gofoom/render/state"
)

func Ceiling(s *state.Slice) {
	mat := material.For(s.Sector.CeilMaterial, s)

	world := &concepts.Vector3{0, 0, s.Sector.TopZ}

	for s.Y = s.YStart; s.Y < s.ClippedStart; s.Y++ {
		if s.Y-s.ScreenHeight/2 == 0 {
			continue
		}

		distToCeil := (s.Sector.TopZ - s.CameraZ) * s.ViewFix[s.X] / float64(s.ScreenHeight/2-1-s.Y)
		scaler := s.Sector.CeilScale / distToCeil
		screenIndex := uint(s.TargetX + s.Y*s.WorkerWidth)

		if distToCeil >= s.ZBuffer[screenIndex] {
			continue
		}

		world.X = s.Map.Player.Pos.X + s.AngleCos*distToCeil
		world.Y = s.Map.Player.Pos.Y + s.AngleSin*distToCeil

		tx := world.X / s.Sector.CeilScale
		tx -= math.Floor(tx)
		ty := world.Y / s.Sector.CeilScale
		ty -= math.Floor(ty)
		tx = math.Abs(tx)
		ty = math.Abs(ty)
		// var light = this.map.light(world, CEIL_NORMAL, slice.sector, slice.segment, null, null, true);

		if mat != nil {
			s.Write(screenIndex, mat.Sample(tx, ty, nil, scaler))
		}
		s.ZBuffer[screenIndex] = distToCeil
	}
}
