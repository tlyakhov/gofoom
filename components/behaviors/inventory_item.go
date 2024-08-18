// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"tlyakhov/gofoom/ecs"
)

type InventoryItem struct {
	ecs.Attached `editable:"^"`

	Class string     `editable:"Class"`
	Count int        `editable:"Count"`
	Image ecs.Entity `editable:"Image" edit_type:"Material"`
}

var InventoryItemComponentIndex int

func init() {
	InventoryItemComponentIndex = ecs.Types().Register(InventoryItem{}, InventoryItemFromDb)
}

func InventoryItemFromDb(db *ecs.ECS, e ecs.Entity) *InventoryItem {
	if asserted, ok := db.Component(e, InventoryItemComponentIndex).(*InventoryItem); ok {
		return asserted
	}
	return nil
}

func (item *InventoryItem) String() string {
	return "InventoryItem"
}

func (item *InventoryItem) Construct(data map[string]any) {
	item.Attached.Construct(data)
	item.Class = "GenericItem"
	item.Count = 1

	if data == nil {
		return
	}

	if v, ok := data["Class"]; ok {
		item.Class = v.(string)
	}
	if v, ok := data["Count"]; ok {
		item.Count = v.(int)
	}
	if v, ok := data["Image"]; ok {
		item.Image, _ = ecs.ParseEntity(v.(string))
	}
}

func (item *InventoryItem) Serialize() map[string]any {
	result := item.Attached.Serialize()

	if item.Class != "GenericItem" {
		result["Class"] = item.Class
	}
	if item.Count > 1 {
		result["Count"] = item.Count
	}
	if item.Image != 0 {
		result["Image"] = item.Image.Format()
	}

	return result
}
