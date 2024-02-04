package behaviors

import (
	"tlyakhov/gofoom/concepts"
)

type InventoryItem struct {
	concepts.Attached `editable:"^"`

	Class string `editable:"Class"`
	Count int    `editable:"Count"`
}

var InventoryItemComponentIndex int

func init() {
	InventoryItemComponentIndex = concepts.DbTypes().Register(InventoryItem{}, InventoryItemFromDb)
}

func InventoryItemFromDb(entity *concepts.EntityRef) *InventoryItem {
	if asserted, ok := entity.Component(InventoryItemComponentIndex).(*InventoryItem); ok {
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
}

func (item *InventoryItem) Serialize() map[string]any {
	result := item.Attached.Serialize()

	if item.Class != "GenericItem" {
		result["Class"] = item.Class
	}
	if item.Count > 1 {
		result["Count"] = item.Count
	}

	return result
}
