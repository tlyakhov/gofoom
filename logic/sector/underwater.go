package sector

import (
	"image/color"

	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/core"
	"tlyakhov/gofoom/entities"
	"tlyakhov/gofoom/sectors"
)

type UnderwaterService struct {
	*PhysicalSectorService
	*sectors.Underwater
}

func NewUnderwaterService(s *sectors.Underwater) *UnderwaterService {
	return &UnderwaterService{Underwater: s, PhysicalSectorService: NewPhysicalSectorService(&s.PhysicalSector)}
}

func (s *UnderwaterService) ActOnEntity(e core.AbstractEntity) {
	if e.GetSector() == nil || e.GetSector().GetBase().ID != s.PhysicalSectorService.ID {
		return
	}

	e.Physical().Vel = e.Physical().Vel.Mul(1.0 / constants.SwimDamping)
	e.Physical().Vel.Z -= constants.GravitySwim

	//if _, ok := e.(*LightEntity); ok {
	//	return
	//}
	s.Collide(e)
}

func (s *UnderwaterService) OnEnter(e core.AbstractEntity) {
	s.PhysicalSectorService.OnEnter(e)
	if p, ok := e.(*entities.Player); ok {
		p.FrameTint = color.NRGBA{75, 147, 255, 90}
	}
}

func (s *UnderwaterService) OnExit(e core.AbstractEntity) {
	s.PhysicalSectorService.OnExit(e)
	if p, ok := e.(*entities.Player); ok {
		p.FrameTint = color.NRGBA{}
	}
}
