package sector

import (
	"github.com/tlyakhov/gofoom/logic/provide"
	"github.com/tlyakhov/gofoom/mapping"
)

type ToxicSectorService struct {
	*PhysicalSectorService
	*mapping.ToxicSector
}

func NewToxicSectorService(s *mapping.ToxicSector) *ToxicSectorService {
	return &ToxicSectorService{ToxicSector: s, PhysicalSectorService: NewPhysicalSectorService(&s.PhysicalSector)}
}

func (s *ToxicSectorService) Collide(e mapping.AbstractEntity) {
	provide.Passer.For(s.ToxicSector.PhysicalSector).Collide(e)
	ae, ok := e.(*mapping.AliveEntity)
	if ok && s.Hurt != 0 && ae.HurtTime == 0 {
		//NewAliveEntityService(ae).Hurt(s.Hurt)
		// TODO: Fix!
	}

	concrete := s.ToxicSector

	if concrete.FloorMaterial != nil && e.GetPhysical().Pos.Z <= concrete.BottomZ {

		//m := registry.Translate(s.FloorMaterial, "logic").(*Painful)
		//m.ActOnEntity(e)
	}
	if concrete.CeilMaterial != nil && e.GetPhysical().Pos.Z >= concrete.TopZ {
		//m := registry.Translate(s.CeilMaterial, "logic").(*Painful)
		//m.ActOnEntity(e)
	}
}
