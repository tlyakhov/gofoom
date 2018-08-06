package sector

import (
	"github.com/tlyakhov/gofoom/core"
	"github.com/tlyakhov/gofoom/entities"
	"github.com/tlyakhov/gofoom/logic/provide"
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
	provide.Passer.For(s.ToxicSector.PhysicalSector).Collide(e)
	ae, ok := e.(*entities.AliveEntity)
	if ok && s.Hurt != 0 && ae.HurtTime == 0 {
		provide.Hurter.For(e).Hurt(s.Hurt)
	}

	concrete := s.ToxicSector

	if concrete.FloorMaterial != nil && e.Physical().Pos.Z <= concrete.BottomZ {
		//m := registry.Translate(s.FloorMaterial, "logic").(*Painful)
		//m.ActOnEntity(e)
	}
	if concrete.CeilMaterial != nil && e.Physical().Pos.Z >= concrete.TopZ {
		//m := registry.Translate(s.CeilMaterial, "logic").(*Painful)
		//m.ActOnEntity(e)
	}
}
