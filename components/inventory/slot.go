// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package inventory

import (
	"strconv"
	"tlyakhov/gofoom/dynamic"
	"tlyakhov/gofoom/ecs"

	"github.com/spf13/cast"
)

// Represents a slot in an inventory that can be filled with an item,
// a weapon, ammo, or other resource.
type Slot struct {
	ecs.Attached `editable:"^"`

	// TODO: Give this some physical structure (e.g. an inventory grid?)
	// TODO: Maybe items could also have mass/weight and player could have a
	// limit? Would enable some RPG mechanics.

	// Should we just use ecs.Named for this?
	Class string `editable:"Class"`

	Limit int                  `editable:"Limit"`
	Count dynamic.Spawned[int] `editable:"Count"`
	Image ecs.Entity           `editable:"Image" edit_type:"Material"`

	Carrier *Carrier
}

var SlotCID ecs.ComponentID

func init() {
	SlotCID = ecs.RegisterComponent(&ecs.Arena[Slot, *Slot]{})
}

func (x *Slot) ComponentID() ecs.ComponentID {
	return SlotCID
}
func GetSlot(e ecs.Entity) *Slot {
	if asserted, ok := ecs.Component(e, SlotCID).(*Slot); ok {
		return asserted
	}
	return nil
}

func (slot *Slot) String() string {
	return "Slot " + slot.Class
}

func (s *Slot) Construct(data map[string]any) {
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

func (s *Slot) IsSystem() bool {
	return false
}

func (s *Slot) Serialize() map[string]any {
	data := s.Attached.Serialize()

	data["Limit"] = strconv.Itoa(s.Limit)
	data["Count"] = s.Count.Serialize()
	data["Class"] = s.Class
	if s.Image != 0 {
		data["Image"] = s.Image.Serialize()
	}

	return data
}
