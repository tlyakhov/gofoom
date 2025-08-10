// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package inventory

import (
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/dynamic"
	"tlyakhov/gofoom/ecs"

	"github.com/spf13/cast"
)

//go:generate go run github.com/dmarkham/enumer -type=ItemFlags -json
type ItemFlags int

const (
	ItemBounce ItemFlags = 1 << iota
	ItemAutoProximity
	ItemAutoPlayerTargetable
)

type Item struct {
	ecs.Attached `editable:"^"`

	Class string               `editable:"Class"`
	Count dynamic.Spawned[int] `editable:"Count"`
	Image ecs.Entity           `editable:"Image" edit_type:"Material"`
	Flags ItemFlags            `editable:"Flags" edit_type:"Flags"`
}

func (item *Item) MultiAttachable() bool { return true }

func (item *Item) String() string {
	return "Item (" + item.Class + ")"
}

func (item *Item) Construct(data map[string]any) {
	item.Attached.Construct(data)
	item.Class = "GenericItem"
	item.Count.SetAll(1)
	item.Flags = ItemBounce | ItemAutoProximity | ItemAutoPlayerTargetable

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
		item.Flags = concepts.ParseFlags(cast.ToString(v), ItemFlagsString)
	}
}

func (item *Item) Serialize() map[string]any {
	result := item.Attached.Serialize()

	result["Count"] = item.Count.Serialize()

	if item.Class != "GenericItem" {
		result["Class"] = item.Class
	}
	if item.Image != 0 {
		result["Image"] = item.Image.Serialize()
	}
	if item.Flags != ItemBounce|ItemAutoProximity|ItemAutoPlayerTargetable {
		result["Flags"] = concepts.SerializeFlags(item.Flags, ItemFlagsValues())
	}

	return result
}
