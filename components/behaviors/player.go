package behaviors

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/sectors"
	"tlyakhov/gofoom/concepts"
)

type Player struct {
	concepts.Attached `editable:"^"`

	FrameTint concepts.Vector4
	Crouching bool
	Inventory []*InventoryItem
	Bob       float64
}

var PlayerComponentIndex int

func init() {
	PlayerComponentIndex = concepts.DbTypes().Register(Player{}, PlayerFromDb)
}

func PlayerFromDb(entity *concepts.EntityRef) *Player {
	if asserted, ok := entity.Component(PlayerComponentIndex).(*Player); ok {
		return asserted
	}
	return nil
}

func (p *Player) Underwater() bool {
	if b := core.BodyFromDb(p.EntityRef); b != nil {
		if u := sectors.UnderwaterFromDb(b.SectorEntityRef); u != nil {
			return true
		}
	}
	return false
}
