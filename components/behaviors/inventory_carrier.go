// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"strconv"
	"tlyakhov/gofoom/ecs"
)

type InventoryCarrier struct {
	ecs.Attached `editable:"^"`

	Inventory     []*InventorySlot `editable:"Inventory"`
	CurrentWeapon *InventorySlot
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

func (ic *InventoryCarrier) Construct(data map[string]any) {
	ic.Attached.Construct(data)

	if data == nil {
		return
	}

	if v, ok := data["Inventory"]; ok {
		ic.Inventory = ecs.ConstructSlice[*InventorySlot](ic.ECS, v, nil)
	}

	if v, ok := data["CurrentWeapon"]; ok {
		index, _ := strconv.Atoi(v.(string))
		if index >= 0 && index < len(ic.Inventory) {
			ic.CurrentWeapon = ic.Inventory[index]
		}
	}
}

func (ic *InventoryCarrier) Serialize() map[string]any {
	result := ic.Attached.Serialize()

	if len(ic.Inventory) > 0 {
		result["Inventory"] = ecs.SerializeSlice(ic.Inventory)
	}
	if ic.CurrentWeapon != nil {
		for i, slot := range ic.Inventory {
			if slot != ic.CurrentWeapon {
				continue
			}
			result["CurrentWeapon"] = strconv.Itoa(i)
			break
		}
	}
	return result
}
