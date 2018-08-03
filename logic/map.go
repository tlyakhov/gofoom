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
	player := registry.Translate(m.Player, "logic").(*Player)
	player.Frame(lastFrameTime)

	for _, item := range m.Sectors {
		sector := item.(*mapping.Sector)
		registry.Translate(sector, "logic").(*Sector).ActOnEntity(registry.Translate(&player.Entity, "logic").(*Entity))
		for _, item2 := range sector.Entities {
			e := registry.Translate(item2, "mapping").(*mapping.Entity)
			if !e.Active {
				continue
			}
			for _, pvs := range sector.PVSEntity {
				_ = pvs
				//pvs.ActOnEntity(e)
			}
		}
		registry.Translate(sector, "logic").(*Sector).Frame(lastFrameTime)
	}
}
