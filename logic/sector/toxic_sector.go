package sector

import (
	"tlyakhov/gofoom/core"
	"tlyakhov/gofoom/logic/material"
	"tlyakhov/gofoom/logic/provide"
	"tlyakhov/gofoom/materials"
	"tlyakhov/gofoom/sectors"
)

type ToxicSectorService struct {
	*PhysicalSectorService
	*sectors.ToxicSector
}

func NewToxicSectorService(s *sectors.ToxicSector) *ToxicSectorService {
	return &ToxicSectorService{ToxicSector: s, PhysicalSectorService: NewPhysicalSectorService(&s.PhysicalSector)}
}

func (s *ToxicSectorService) Collide(e core.AbstractEntity) {
	if e.GetSector() == nil || e.GetSector().GetBase().ID != s.PhysicalSectorService.ID {
		return
	}

	if h, ok := provide.Hurter.For(e); ok && h.HurtTime() == 0 && s.Hurt > 0 {
		h.Hurt(s.Hurt)
	}

	concrete := s.ToxicSector

	if concrete.FloorMaterial != nil && e.Physical().Pos[2] <= concrete.BottomZ {
		if p, ok := concrete.FloorMaterial.(*materials.PainfulLitSampled); ok {
			material.NewPainfulService(&p.Painful).ActOnEntity(e)
		}
	}
	if concrete.CeilMaterial != nil && e.Physical().Pos[2] >= concrete.TopZ {
		if p, ok := concrete.CeilMaterial.(*materials.PainfulLitSampled); ok {
			material.NewPainfulService(&p.Painful).ActOnEntity(e)
		}
	}
}

func (s *ToxicSectorService) ActOnEntity(e core.AbstractEntity) {
	s.Collide(e)
	s.PhysicalSectorService.ActOnEntity(e)
}
