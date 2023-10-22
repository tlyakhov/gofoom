package sector

import (
	"tlyakhov/gofoom/controllers/material"
	"tlyakhov/gofoom/controllers/provide"
	"tlyakhov/gofoom/core"
	"tlyakhov/gofoom/materials"
	"tlyakhov/gofoom/sectors"
)

type ToxicSectorController struct {
	*PhysicalSectorController
	*sectors.ToxicSector
}

func NewToxicSectorController(s *sectors.ToxicSector) *ToxicSectorController {
	return &ToxicSectorController{ToxicSector: s, PhysicalSectorController: NewPhysicalSectorController(&s.PhysicalSector)}
}

func (s *ToxicSectorController) Collide(e core.AbstractMob) {
	if e.GetSector() == nil || e.GetSector().GetBase().ID != s.PhysicalSectorController.ID {
		return
	}

	if h := provide.Hurter.For(e); h != nil && h.HurtTime() == 0 && s.Hurt > 0 {
		h.Hurt(s.Hurt)
	}

	model := s.ToxicSector

	if model.FloorMaterial != nil && e.Physical().Pos.Now[2] <= model.BottomZ.Now {
		if p, ok := model.FloorMaterial.(*materials.PainfulLitSampled); ok {
			material.NewPainfulController(&p.Painful).ActOnMob(e)
		}
	}
	if model.CeilMaterial != nil && e.Physical().Pos.Now[2] >= model.TopZ.Now {
		if p, ok := model.CeilMaterial.(*materials.PainfulLitSampled); ok {
			material.NewPainfulController(&p.Painful).ActOnMob(e)
		}
	}
}

func (s *ToxicSectorController) ActOnMob(e core.AbstractMob) {
	s.Collide(e)
	s.PhysicalSectorController.ActOnMob(e)
}
