package sector

import (
	"github.com/tlyakhov/gofoom/constants"
	"github.com/tlyakhov/gofoom/core"
	"github.com/tlyakhov/gofoom/logic/provide"
)

type PhysicalSectorService struct {
	*core.PhysicalSector
}

func NewPhysicalSectorService(s *core.PhysicalSector) *PhysicalSectorService {
	return &PhysicalSectorService{PhysicalSector: s}
}

func (s *PhysicalSectorService) OnEnter(e core.AbstractEntity) {
	if s.FloorTarget == nil && e.Physical().Pos.Z <= e.GetSector().Physical().BottomZ {
		e.Physical().Pos.Z = e.GetSector().Physical().BottomZ
	}
}

func (s *PhysicalSectorService) OnExit(e core.AbstractEntity) {
}

func (s *PhysicalSectorService) Collide(e core.AbstractEntity) {
	concrete := e.Physical()
	entityTop := concrete.Pos.Z + concrete.Height

	if s.FloorTarget != nil && entityTop < s.BottomZ {
		provide.Passer.For(e.GetSector()).OnExit(e)
		concrete.Sector = s.FloorTarget
		provide.Passer.For(e.GetSector()).OnEnter(e)
		concrete.Pos.Z = e.GetSector().Physical().TopZ - concrete.Height - 1.0
	} else if s.FloorTarget != nil && concrete.Pos.Z <= s.BottomZ && concrete.Vel.Z > 0 {
		concrete.Vel.Z = constants.PlayerJumpStrength
	} else if s.FloorTarget == nil && concrete.Pos.Z <= s.BottomZ {
		concrete.Vel.Z = 0
		concrete.Pos.Z = s.BottomZ
	}

	if s.CeilTarget != nil && entityTop > s.TopZ {
		provide.Passer.For(e.GetSector()).OnExit(e)
		concrete.Sector = s.CeilTarget
		provide.Passer.For(e.GetSector()).OnEnter(e)
		concrete.Pos.Z = e.GetSector().Physical().BottomZ - concrete.Height + 1.0
	} else if s.CeilTarget == nil && entityTop > s.TopZ {
		concrete.Vel.Z = 0
		concrete.Pos.Z = s.TopZ - concrete.Height - 1.0
	}
}

func (s *PhysicalSectorService) ActOnEntity(e core.AbstractEntity) {
	if e.GetSector() == nil || e.GetSector().GetBase().ID != s.ID {
		return
	}

	if e.GetBase().ID == s.Map.Player.GetBase().ID {
		e.Physical().Vel.X = 0
		e.Physical().Vel.Y = 0
	}

	e.Physical().Vel.Z -= constants.Gravity

	s.Collide(e)
}

func (s *PhysicalSectorService) Frame(lastFrameTime float64) {
	for _, e := range s.Entities {
		if e.GetBase().ID == s.Map.Player.GetBase().ID || s.Map.EntitiesPaused {
			continue
		}
		provide.EntityAnimator.For(e).Frame(lastFrameTime)
	}
}
