// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"strconv"
	"strings"
	"tlyakhov/gofoom/containers"
	"tlyakhov/gofoom/dynamic"
	"tlyakhov/gofoom/ecs"

	"github.com/spf13/cast"
)

type InventorySlot struct {
	ECS          *ecs.ECS
	ValidClasses containers.Set[string] `editable:"Classes"`
	Limit        int                    `editable:"Limit"`
	Count        dynamic.Spawned[int]   `editable:"Count"`
	Image        ecs.Entity             `editable:"Image" edit_type:"Material"`

	// TODO: Instead, in the future, maybe it makes sense to have inventory slots
	// be their own entities, and just hang WeaponInstant (or similar) function
	// components off of those.
	WeaponState ecs.Entity
}

func (s *InventorySlot) Construct(data map[string]any) {
	s.Limit = 100
	s.Count.SetAll(0)
	s.Image = 0
	s.ValidClasses = make(containers.Set[string])

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
	if v, ok := data["State"]; ok {
		s.WeaponState, _ = ecs.ParseEntity(v.(string))
	}
	if v, ok := data["ValidClasses"]; ok {
		classes := strings.Split(v.(string), ",")
		for _, c := range classes {
			s.ValidClasses.Add(strings.TrimSpace(c))
		}
	}
}

func (s *InventorySlot) IsSystem() bool {
	return false
}

func (s *InventorySlot) Serialize() map[string]any {
	data := make(map[string]any)
	data["Limit"] = strconv.Itoa(s.Limit)
	data["Count"] = s.Count.Serialize()
	if s.Image != 0 {
		data["Image"] = s.Image.String()
	}
	if s.WeaponState != 0 {
		data["WeaponState"] = s.WeaponState.String()
	}
	classes := ""
	for c := range s.ValidClasses {
		if len(classes) > 0 {
			classes += ","
		}
		classes += c
	}
	data["ValidClasses"] = classes
	return data
}

func (s *InventorySlot) OnAttach(db *ecs.ECS) {
	s.ECS = db
}
func (s *InventorySlot) GetECS() *ecs.ECS {
	return s.ECS
}
