// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package inventory

import (
	"strings"
	"tlyakhov/gofoom/ecs"

	"github.com/spf13/cast"
)

// Can be attached to either a player or an NPC
type Carrier struct {
	ecs.Attached `editable:"^"`

	Slots ecs.EntityTable `editable:"Slots" edit_type:"InventorySlot"`
	// TODO: Add flexibility, e.g. dual wielding, alts, etc...
	SelectedWeapon ecs.Entity
}

func (c *Carrier) String() string {
	if len(c.Slots) == 0 {
		return "Empty Carrier"
	}
	return "Carrier " + c.Slots.String()
}

func (c *Carrier) Shareable() bool {
	return true
}

func (c *Carrier) OnDetach(e ecs.Entity) {
	defer c.Attached.OnDetach(e)
	if !c.IsAttached() {
		return
	}
	for _, slot := range c.Slots {
		// Slots are not shareable and therefore unique to their
		// carriers. Spawners will copy them for players.
		// Don't need to check for zero, Delete will ignore it.
		ecs.Delete(slot)
	}
}

func (c *Carrier) HasAtLeast(class string, min int) bool {
	for _, e := range c.Slots {
		if e == 0 {
			continue
		}
		if slot := GetSlot(e); slot != nil {
			// log.Printf("HasAtLeast %v, %v ? %v, %v", class, min, slot.Class, slot.Count.Now)
			if slot.Class == strings.TrimSpace(class) && slot.Count.Now >= min {
				return true
			}
		}
	}
	return false
}

func (c *Carrier) Construct(data map[string]any) {
	c.Attached.Construct(data)

	if data == nil {
		return
	}

	if v, ok := data["Inventory"]; ok {
		c.Slots = ecs.ParseEntityTable(v, false)
	}

	if v, ok := data["SelectedWeapon"]; ok {
		c.SelectedWeapon, _ = ecs.ParseEntity(cast.ToString(v))
	}
}

func (c *Carrier) Serialize() map[string]any {
	result := c.Attached.Serialize()

	result["Inventory"] = c.Slots.Serialize()

	if c.SelectedWeapon != 0 {
		result["SelectedWeapon"] = c.SelectedWeapon.Serialize()
	}

	return result
}
