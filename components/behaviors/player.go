package behaviors

import (
	"image/color"

	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
)

type Player struct {
	concepts.Attached `editable:"^"`

	FrameTint color.NRGBA
	Crouching bool
	Inventory []core.Body
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
