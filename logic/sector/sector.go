package sector

import (
	"github.com/tlyakhov/gofoom/constants"
	"github.com/tlyakhov/gofoom/logic/provide"
	"github.com/tlyakhov/gofoom/mapping"
)

type SectorService struct {
	*mapping.Sector
}

func NewSectorService(s *mapping.Sector) *SectorService {
	return &SectorService{Sector: s}
}

func (s *SectorService) OnEnter(e mapping.AbstractEntity) {
	if s.FloorTarget == nil && e.GetEntity().Pos.Z <= e.GetEntity().Sector.GetSector().BottomZ {
		e.GetEntity().Pos.Z = e.GetEntity().Sector.GetSector().BottomZ
	}
}

func (s *SectorService) OnExit(e mapping.AbstractEntity) {
}

func (s *SectorService) Collide(e mapping.AbstractEntity) {
	concrete := e.GetEntity()
	entityTop := concrete.Pos.Z + concrete.Height

	if s.FloorTarget != nil && entityTop < s.BottomZ {
		provide.Passer.For(e.GetSector()).OnExit(e)
		concrete.Sector = s.FloorTarget
		provide.Passer.For(e.GetSector()).OnEnter(e)
		concrete.Pos.Z = e.GetSector().GetSector().TopZ - concrete.Height - 1.0
	} else if s.FloorTarget == nil && concrete.Pos.Z <= s.BottomZ {
		concrete.Vel.Z = 0
		concrete.Pos.Z = s.BottomZ
	}

	if s.CeilTarget != nil && entityTop > s.TopZ {
		provide.Passer.For(e.GetSector()).OnExit(e)
		concrete.Sector = s.CeilTarget
		provide.Passer.For(e.GetSector()).OnEnter(e)
		concrete.Pos.Z = e.GetSector().GetSector().BottomZ - concrete.Height + 1.0
	} else if s.CeilTarget == nil && entityTop > s.TopZ {
		concrete.Vel.Z = 0
		concrete.Pos.Z = s.TopZ - concrete.Height - 1.0
	}
}

func (s *SectorService) ActOnEntity(e mapping.AbstractEntity) {
	if e.GetSector() == nil || e.GetSector().GetBase().ID != s.ID {
		return
	}

	if e.GetBase().ID == s.Map.Player.ID {
		e.GetEntity().Vel.X = 0
		e.GetEntity().Vel.Y = 0
	}

	e.GetEntity().Vel.Z -= constants.Gravity

	s.Collide(e)
}

func (s *SectorService) Frame(lastFrameTime float64) {
	for _, e := range s.Entities {
		if e.GetBase().ID == s.Map.Player.ID || s.Map.EntitiesPaused {
			continue
		}
		provide.EntityAnimator.For(e).Frame(lastFrameTime)
	}
}
