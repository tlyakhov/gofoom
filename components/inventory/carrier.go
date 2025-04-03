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

var CarrierCID ecs.ComponentID

func init() {
	CarrierCID = ecs.RegisterComponent(&ecs.Column[Carrier, *Carrier]{Getter: GetCarrier})
}

func (*Carrier) ComponentID() ecs.ComponentID {
	return CarrierCID
}
func GetCarrier(u *ecs.Universe, e ecs.Entity) *Carrier {
	if asserted, ok := u.Component(e, CarrierCID).(*Carrier); ok {
		return asserted
	}
	return nil
}

func (c *Carrier) String() string {
	return "Carrier: " + c.Slots.String()
}

func (c *Carrier) MultiAttachable() bool {
	return true
}

func (c *Carrier) HasAtLeast(class string, min int) bool {
	for _, e := range c.Slots {
		if e == 0 {
			continue
		}
		if slot := GetSlot(c.Universe, e); slot != nil {
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

	result["Inventory"] = c.Slots.Serialize(c.Universe)

	if c.SelectedWeapon != 0 {
		result["SelectedWeapon"] = c.SelectedWeapon.Serialize(c.Universe)
	}

	return result
}
