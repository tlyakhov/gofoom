// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"tlyakhov/gofoom/dynamic"
	"tlyakhov/gofoom/ecs"
)

type InventoryItem struct {
	ecs.Attached `editable:"^"`

	Class string               `editable:"Class"`
	Count dynamic.Spawned[int] `editable:"Count"`
	Image ecs.Entity           `editable:"Image" edit_type:"Material"`
}

var InventoryItemCID ecs.ComponentID

func init() {
	InventoryItemCID = ecs.RegisterComponent(&ecs.Column[InventoryItem, *InventoryItem]{Getter: GetInventoryItem})
}

func GetInventoryItem(db *ecs.ECS, e ecs.Entity) *InventoryItem {
	if asserted, ok := db.Component(e, InventoryItemCID).(*InventoryItem); ok {
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
	item.Count.SetAll(1)

	if data == nil {
		return
	}

	if v, ok := data["Class"]; ok {
		item.Class = v.(string)
	}
	if v, ok := data["Count"]; ok {
		item.Count.Construct(v.(map[string]any))
	}
	if v, ok := data["Image"]; ok {
		item.Image, _ = ecs.ParseEntity(v.(string))
	}
}

func (item *InventoryItem) Serialize() map[string]any {
	result := item.Attached.Serialize()

	result["Count"] = item.Count.Serialize()

	if item.Class != "GenericItem" {
		result["Class"] = item.Class
	}
	if item.Image != 0 {
		result["Image"] = item.Image.String()
	}

	return result
}
