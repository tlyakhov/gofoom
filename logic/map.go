package logic

import (
	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/mapping"
)

type Map mapping.Map

func (m *Map) Frame(lastFrameTime float64) {
	concepts.Local(m.Player, TypeMap).(*Player).Frame(lastFrameTime)

	for _, item := range m.Sectors {
		sector := item.(*mapping.Sector)
		// sector.ActOnEntity(player)
		for _, item2 := range sector.Entities {
			e := item2.(*mapping.Entity)
			if !e.Active {
				continue
			}
			for _, pvs := range sector.PVSEntity {
				_ = pvs
				//pvs.ActOnEntity(e)
			}
		}
		concepts.Local(sector, TypeMap).(*Sector).Frame(lastFrameTime)
	}
}
