package sector

import (
	"github.com/tlyakhov/gofoom/constants"
	"github.com/tlyakhov/gofoom/logic/provide"
	"github.com/tlyakhov/gofoom/mapping"
)

type PhysicalSectorService struct {
	*mapping.PhysicalSector
}

func NewPhysicalSectorService(s *mapping.PhysicalSector) *PhysicalSectorService {
	return &PhysicalSectorService{PhysicalSector: s}
}

func (s *PhysicalSectorService) OnEnter(e mapping.AbstractEntity) {
	if s.FloorTarget == nil && e.GetPhysical().Pos.Z <= e.GetSector().GetPhysical().BottomZ {
		e.GetPhysical().Pos.Z = e.GetSector().GetPhysical().BottomZ
	}
}

func (s *PhysicalSectorService) OnExit(e mapping.AbstractEntity) {
}

func (s *PhysicalSectorService) Collide(e mapping.AbstractEntity) {
	concrete := e.GetPhysical()
	entityTop := concrete.Pos.Z + concrete.Height

	if s.FloorTarget != nil && entityTop < s.BottomZ {
		provide.Passer.For(e.GetSector()).OnExit(e)
		concrete.Sector = s.FloorTarget
		provide.Passer.For(e.GetSector()).OnEnter(e)
		concrete.Pos.Z = e.GetSector().GetPhysical().TopZ - concrete.Height - 1.0
	} else if s.FloorTarget == nil && concrete.Pos.Z <= s.BottomZ {
		concrete.Vel.Z = 0
		concrete.Pos.Z = s.BottomZ
	}

	if s.CeilTarget != nil && entityTop > s.TopZ {
		provide.Passer.For(e.GetSector()).OnExit(e)
		concrete.Sector = s.CeilTarget
		provide.Passer.For(e.GetSector()).OnEnter(e)
		concrete.Pos.Z = e.GetSector().GetPhysical().BottomZ - concrete.Height + 1.0
	} else if s.CeilTarget == nil && entityTop > s.TopZ {
		concrete.Vel.Z = 0
		concrete.Pos.Z = s.TopZ - concrete.Height - 1.0
	}
}

func (s *PhysicalSectorService) ActOnEntity(e mapping.AbstractEntity) {
	if e.GetSector() == nil || e.GetSector().GetBase().ID != s.ID {
		return
	}

	if e.GetBase().ID == s.Map.Player.ID {
		e.GetPhysical().Vel.X = 0
		e.GetPhysical().Vel.Y = 0
	}

	e.GetPhysical().Vel.Z -= constants.Gravity

	s.Collide(e)
}

func (s *PhysicalSectorService) Frame(lastFrameTime float64) {
	for _, e := range s.Entities {
		if e.GetBase().ID == s.Map.Player.ID || s.Map.EntitiesPaused {
			continue
		}
		provide.EntityAnimator.For(e).Frame(lastFrameTime)
	}
}
