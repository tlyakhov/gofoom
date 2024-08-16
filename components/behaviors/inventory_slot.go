// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"strconv"
	"strings"
	"tlyakhov/gofoom/concepts"
)

type InventorySlot struct {
	DB           *concepts.EntityComponentDB
	ValidClasses concepts.Set[string]
	Limit        int
	Count        int
	Image        concepts.Entity `editable:"Image" edit_type:"Material"`
}

func (s *InventorySlot) Construct(data map[string]any) {
	s.Limit = 100
	s.Count = 0
	s.Image = 0
	s.ValidClasses = make(concepts.Set[string])

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
		s.Image, _ = concepts.ParseEntity(v.(string))
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
	data["Image"] = s.Image.Format()
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

func (s *InventorySlot) SetDB(db *concepts.EntityComponentDB) {
	s.DB = db
}
func (s *InventorySlot) GetDB() *concepts.EntityComponentDB {
	return s.DB
}
