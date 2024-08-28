// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"strconv"
	"strings"
	"tlyakhov/gofoom/containers"
	"tlyakhov/gofoom/ecs"
)

type InventorySlot struct {
	ECS          *ecs.ECS
	ValidClasses containers.Set[string] `editable:"Classes"`
	Limit        int                    `editable:"Limit"`
	Count        int                    `editable:"Count"`
	Image        ecs.Entity             `editable:"Image" edit_type:"Material"`
}

func (s *InventorySlot) Construct(data map[string]any) {
	s.Limit = 100
	s.Count = 0
	s.Image = 0
	s.ValidClasses = make(containers.Set[string])

	if data == nil {
		return
	}

	if v, ok := data["Limit"]; ok {
		s.Limit, _ = strconv.Atoi(v.(string))
	}
	if v, ok := data["Count"]; ok {
		s.Count, _ = strconv.Atoi(v.(string))
	}
	if v, ok := data["Image"]; ok {
		s.Image, _ = ecs.ParseEntity(v.(string))
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
	data["Count"] = strconv.Itoa(s.Count)
	data["Image"] = s.Image.String()
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

func (s *InventorySlot) AttachECS(db *ecs.ECS) {
	s.ECS = db
}
func (s *InventorySlot) GetECS() *ecs.ECS {
	return s.ECS
}
