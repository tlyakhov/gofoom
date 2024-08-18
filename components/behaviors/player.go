// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"strconv"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/ecs"
)

type Player struct {
	ecs.Attached `editable:"^"`

	Spawn         bool
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
	PlayerComponentIndex = ecs.Types().Register(Player{}, PlayerFromDb)
}

func PlayerFromDb(db *ecs.ECS, e ecs.Entity) *Player {
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
	// By convention, we construct spawn points rather than active players to
	// avoid weird behaviors.
	p.Spawn = true

	if data == nil {
		return
	}

	if v, ok := data["Spawn"]; ok {
		p.Spawn = v.(bool)
	}

	if v, ok := data["Inventory"]; ok {
		ecs.ConstructSlice[*InventoryItem](p.DB, v)
	}

	if v, ok := data["CurrentWeapon"]; ok {
		index, _ := strconv.Atoi(v.(string))
		if index >= 0 && index < len(p.Inventory) {
			p.CurrentWeapon = p.Inventory[index]
		}
	}
}

func (p *Player) Serialize() map[string]any {
	result := p.Attached.Serialize()

	result["Spawn"] = p.Spawn

	if len(p.Inventory) > 0 {
		result["Inventory"] = ecs.SerializeSlice(p.Inventory)
	}
	if p.CurrentWeapon != nil {
		for i, slot := range p.Inventory {
			if slot != p.CurrentWeapon {
				continue
			}
			result["CurrentWeapon"] = strconv.Itoa(i)
			break
		}
	}
	return result
}
