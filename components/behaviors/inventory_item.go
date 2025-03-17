// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/dynamic"
	"tlyakhov/gofoom/ecs"

	"github.com/spf13/cast"
)

//go:generate go run github.com/dmarkham/enumer -type=InventoryItemFlags -json
type InventoryItemFlags int

const (
	InventoryItemBounce InventoryItemFlags = 1 << iota
	InventoryItemAutoProximity
	InventoryItemAutoPlayerTargetable
)

type InventoryItem struct {
	ecs.Attached `editable:"^"`

	Class string               `editable:"Class"`
	Count dynamic.Spawned[int] `editable:"Count"`
	Image ecs.Entity           `editable:"Image" edit_type:"Material"`
	Flags InventoryItemFlags   `editable:"Flags" edit_type:"Flags"`
}

var InventoryItemCID ecs.ComponentID

func init() {
	InventoryItemCID = ecs.RegisterComponent(&ecs.Column[InventoryItem, *InventoryItem]{Getter: GetInventoryItem})
}

func GetInventoryItem(u *ecs.Universe, e ecs.Entity) *InventoryItem {
	if asserted, ok := u.Component(e, InventoryItemCID).(*InventoryItem); ok {
		return asserted
	}
	return nil
}

func (item *InventoryItem) MultiAttachable() bool { return true }

func (item *InventoryItem) String() string {
	return "InventoryItem"
}

func (item *InventoryItem) Construct(data map[string]any) {
	item.Attached.Construct(data)
	item.Class = "GenericItem"
	item.Count.SetAll(1)
	item.Flags = InventoryItemBounce | InventoryItemAutoProximity | InventoryItemAutoPlayerTargetable

	if data == nil {
		return
	}

	if v, ok := data["Class"]; ok {
		item.Class = cast.ToString(v)
	}
	if v, ok := data["Count"]; ok {
		item.Count.Construct(v.(map[string]any))
	}
	if v, ok := data["Image"]; ok {
		item.Image, _ = ecs.ParseEntity(v.(string))
	}
	if v, ok := data["Flags"]; ok {
		item.Flags = concepts.ParseFlags(cast.ToString(v), InventoryItemFlagsString)
	}
}

func (item *InventoryItem) Serialize() map[string]any {
	result := item.Attached.Serialize()

	result["Count"] = item.Count.Serialize()

	if item.Class != "GenericItem" {
		result["Class"] = item.Class
	}
	if item.Image != 0 {
		result["Image"] = item.Image.Serialize(item.Universe)
	}
	if item.Flags != InventoryItemBounce|InventoryItemAutoProximity|InventoryItemAutoPlayerTargetable {
		result["Flags"] = concepts.SerializeFlags(item.Flags, InventoryItemFlagsValues())
	}

	return result
}
