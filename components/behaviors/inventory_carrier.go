// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"tlyakhov/gofoom/ecs"

	"github.com/spf13/cast"
)

// Can be attached to either a player or an NPC
type InventoryCarrier struct {
	ecs.Attached `editable:"^"`

	Inventory ecs.EntityTable `editable:"Inventory" edit_type:"InventorySlot"`
	// TODO: Add flexibility, e.g. dual wielding, alts, etc...
	SelectedWeapon ecs.Entity
}

var InventoryCarrierCID ecs.ComponentID

func init() {
	InventoryCarrierCID = ecs.RegisterComponent(&ecs.Column[InventoryCarrier, *InventoryCarrier]{Getter: GetInventoryCarrier})
}

func GetInventoryCarrier(db *ecs.ECS, e ecs.Entity) *InventoryCarrier {
	if asserted, ok := db.Component(e, InventoryCarrierCID).(*InventoryCarrier); ok {
		return asserted
	}
	return nil
}

func (ic *InventoryCarrier) HasAtLeast(class string, min int) bool {
	for _, e := range ic.Inventory {
		if e == 0 {
			continue
		}
		if slot := GetInventorySlot(ic.ECS, e); slot != nil {
			if slot.Class == class && slot.Count.Now > min {
				return true
			}
		}
	}
	return false
}

func (ic *InventoryCarrier) Construct(data map[string]any) {
	ic.Attached.Construct(data)

	if data == nil {
		return
	}

	if v, ok := data["Inventory"]; ok {
		ic.Inventory = ecs.ParseEntityTable(v)
	}

	if v, ok := data["SelectedWeapon"]; ok {
		ic.SelectedWeapon, _ = ecs.ParseEntity(cast.ToString(v))
	}
}

func (ic *InventoryCarrier) Serialize() map[string]any {
	result := ic.Attached.Serialize()

	result["Inventory"] = ic.Inventory.Serialize()

	if ic.SelectedWeapon != 0 {
		result["SelectedWeapon"] = ic.SelectedWeapon.String()
	}

	return result
}
