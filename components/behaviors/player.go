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
	Inventory []core.Mob
	Bob       float64
}

var PlayerComponentIndex int

func init() {
	PlayerComponentIndex = concepts.DbTypes().Register(Player{})
}

func PlayerFromDb(entity *concepts.EntityRef) *Player {
	return entity.Component(PlayerComponentIndex).(*Player)
}
