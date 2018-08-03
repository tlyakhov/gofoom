package logic

import (
	"github.com/tlyakhov/gofoom/mapping"
	"github.com/tlyakhov/gofoom/registry"
)

type Map mapping.Map

func init() {
	registry.Instance().RegisterMapped(Map{}, mapping.Map{})
}

func (m *Map) Frame(lastFrameTime float64) {
	registry.Translate(m.Player).(*Player).Frame(lastFrameTime)

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
		registry.Translate(sector).(*Sector).Frame(lastFrameTime)
	}
}
