package sector

import (
	"github.com/tlyakhov/gofoom/core"
	"github.com/tlyakhov/gofoom/logic/material"
	"github.com/tlyakhov/gofoom/logic/provide"
	"github.com/tlyakhov/gofoom/materials"
	"github.com/tlyakhov/gofoom/sectors"
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

	if concrete.FloorMaterial != nil && e.Physical().Pos.Z <= concrete.BottomZ {
		if p, ok := concrete.FloorMaterial.(*materials.PainfulLitSampled); ok {
			material.NewPainfulService(&p.Painful).ActOnEntity(e)
		}
	}
	if concrete.CeilMaterial != nil && e.Physical().Pos.Z >= concrete.TopZ {
		if p, ok := concrete.CeilMaterial.(*materials.PainfulLitSampled); ok {
			material.NewPainfulService(&p.Painful).ActOnEntity(e)
		}
	}
}

func (s *ToxicSectorService) ActOnEntity(e core.AbstractEntity) {
	s.Collide(e)
	s.PhysicalSectorService.ActOnEntity(e)
}
