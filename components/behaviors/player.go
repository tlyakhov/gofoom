// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
)

type Player struct {
	concepts.Attached `editable:"^"`

	FrameTint     concepts.Vector4
	Crouching     bool
	Inventory     []*InventorySlot `editable:"Inventory"`
	Bob           float64
	CameraZ       float64
	CurrentWeapon *InventorySlot

	Notices concepts.SyncUniqueQueue
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

func (p *Player) Construct(data map[string]any) {
	p.Attached.Construct(data)

	if data == nil {
		return
	}

	if v, ok := data["Inventory"]; ok {
		concepts.ConstructSlice[*InventoryItem](p.DB, v)
	}

	/*if v, ok := data["CurrentWeapon"]; ok {
		p.CurrentWeapon, _ = concepts.ParseEntity(v.(string))
	}*/
}

func (p *Player) Serialize() map[string]any {
	result := p.Attached.Serialize()

	if len(p.Inventory) > 0 {
		result["Inventory"] = concepts.SerializeSlice(p.Inventory)
	}
	/*if p.CurrentWeapon != 0 {
		result["CurrentWeapon"] = p.CurrentWeapon.Format()
	}*/
	return result
}
