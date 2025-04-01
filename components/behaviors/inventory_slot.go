// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"strconv"
	"tlyakhov/gofoom/dynamic"
	"tlyakhov/gofoom/ecs"

	"github.com/spf13/cast"
)

// Represents a slot in an inventory that can be filled with an item,
// a weapon, ammo, or other resource.
type InventorySlot struct {
	ecs.Attached `editable:"^"`

	// TODO: Give this some physical structure (e.g. an inventory grid)
	// Should we just use ecs.Named for this?
	Class string               `editable:"Class"`
	Limit int                  `editable:"Limit"`
	Count dynamic.Spawned[int] `editable:"Count"`
	Image ecs.Entity           `editable:"Image" edit_type:"Material"`

	Carrier *InventoryCarrier
}

var InventorySlotCID ecs.ComponentID

func init() {
	InventorySlotCID = ecs.RegisterComponent(&ecs.Column[InventorySlot, *InventorySlot]{Getter: GetInventorySlot})
}

func (x *InventorySlot) ComponentID() ecs.ComponentID {
	return InventorySlotCID
}
func GetInventorySlot(u *ecs.Universe, e ecs.Entity) *InventorySlot {
	if asserted, ok := u.Component(e, InventorySlotCID).(*InventorySlot); ok {
		return asserted
	}
	return nil
}

func (slot *InventorySlot) String() string {
	return "InventorySlot"
}

func (s *InventorySlot) Construct(data map[string]any) {
	s.Attached.Construct(data)

	s.Limit = 100
	s.Count.SetAll(0)
	s.Image = 0
	s.Class = ""

	if data == nil {
		return
	}

	if v, ok := data["Limit"]; ok {
		s.Limit = cast.ToInt(v)
	}
	if v, ok := data["Count"]; ok {
		s.Count.Construct(v.(map[string]any))
	}
	if v, ok := data["Image"]; ok {
		s.Image, _ = ecs.ParseEntity(v.(string))
	}
	if v, ok := data["Class"]; ok {
		s.Class = cast.ToString(v)
	}
}

func (s *InventorySlot) IsSystem() bool {
	return false
}

func (s *InventorySlot) Serialize() map[string]any {
	data := s.Attached.Serialize()

	data["Limit"] = strconv.Itoa(s.Limit)
	data["Count"] = s.Count.Serialize()
	data["Class"] = s.Class
	if s.Image != 0 {
		data["Image"] = s.Image.Serialize(s.Universe)
	}

	return data
}
