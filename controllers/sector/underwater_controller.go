package sector

import (
	"image/color"

	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/core"
	"tlyakhov/gofoom/mobs"
	"tlyakhov/gofoom/sectors"
)

type UnderwaterController struct {
	*PhysicalSectorController
	*sectors.Underwater
}

func NewUnderwaterController(s *sectors.Underwater) *UnderwaterController {
	return &UnderwaterController{Underwater: s, PhysicalSectorController: NewPhysicalSectorController(&s.PhysicalSector)}
}

func (s *UnderwaterController) ActOnMob(e core.AbstractMob) {
	if e.GetSector() == nil || e.GetSector().GetBase().Name != s.PhysicalSectorController.Name {
		return
	}

	e.Physical().Vel.Now.MulSelf(1.0 / constants.SwimDamping)
	e.Physical().Vel.Now[2] -= constants.GravitySwim
	s.Collide(e)
}

func (s *UnderwaterController) OnEnter(e core.AbstractMob) {
	s.PhysicalSectorController.OnEnter(e)
	if p, ok := e.(*mobs.Player); ok {
		p.FrameTint = color.NRGBA{75, 147, 255, 90}
	}
}

func (s *UnderwaterController) OnExit(e core.AbstractMob) {
	s.PhysicalSectorController.OnExit(e)
	if p, ok := e.(*mobs.Player); ok {
		p.FrameTint = color.NRGBA{}
	}
}
