package mapping

import (
	"github.com/tlyakhov/gofoom/constants"
)

type Player struct {
	AliveEntity

	Height    float64
	Standing  bool
	Crouching bool
	Inventory []Entity
}

func (p *Player) Initialize() {
	p.AliveEntity.Initialize()
	p.Height = constants.PlayerHeight
	p.BoundingRadius = constants.PlayerBoundingRadius
	p.Standing = true
}
