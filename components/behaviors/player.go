// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
)

type Player struct {
	concepts.Attached `editable:"^"`

	FrameTint concepts.Vector4
	Crouching bool
	Inventory []*InventoryItem
	Bob       float64
}

var PlayerComponentIndex int

func init() {
	PlayerComponentIndex = concepts.DbTypes().Register(Player{}, PlayerFromDb)
}

func PlayerFromDb(db *concepts.EntityComponentDB, e concepts.Entity) *Player {
	if asserted, ok := db.Component(e, PlayerComponentIndex).(*Player); ok {
		return asserted
	}
	return nil
}

func (p *Player) Underwater() bool {
	if b := core.BodyFromDb(p.DB, p.Entity); b != nil {
		if u := UnderwaterFromDb(p.DB, b.SectorEntity); u != nil {
			return true
		}
	}
	return false
}
