package sector

import (
	"image/color"

	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/core"
	"tlyakhov/gofoom/entities"
	"tlyakhov/gofoom/sectors"
)

type UnderwaterController struct {
	*PhysicalSectorController
	*sectors.Underwater
}

func NewUnderwaterController(s *sectors.Underwater) *UnderwaterController {
	return &UnderwaterController{Underwater: s, PhysicalSectorController: NewPhysicalSectorController(&s.PhysicalSector)}
}

func (s *UnderwaterController) ActOnEntity(e core.AbstractEntity) {
	if e.GetSector() == nil || e.GetSector().GetBase().ID != s.PhysicalSectorController.ID {
		return
	}

	e.Physical().Vel.Now.MulSelf(1.0 / constants.SwimDamping)
	e.Physical().Vel.Now[2] -= constants.GravitySwim

	//if _, ok := e.(*LightEntity); ok {
	//	return
	//}
	s.Collide(e)
}

func (s *UnderwaterController) OnEnter(e core.AbstractEntity) {
	s.PhysicalSectorController.OnEnter(e)
	if p, ok := e.(*entities.Player); ok {
		p.FrameTint = color.NRGBA{75, 147, 255, 90}
	}
}

func (s *UnderwaterController) OnExit(e core.AbstractEntity) {
	s.PhysicalSectorController.OnExit(e)
	if p, ok := e.(*entities.Player); ok {
		p.FrameTint = color.NRGBA{}
	}
}
