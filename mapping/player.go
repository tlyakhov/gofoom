package mapping

import (
	"github.com/tlyakhov/gofoom/constants"
	"github.com/tlyakhov/gofoom/registry"
)

type Player struct {
	AliveEntity

	Height    float64
	Standing  bool
	Crouching bool
	Inventory []Entity
}

func init() {
	registry.Instance().Register(Player{})
}

func (p *Player) Initialize() {
	p.AliveEntity.Initialize()
	p.Height = constants.PlayerHeight
	p.BoundingRadius = constants.PlayerBoundingRadius
	p.Standing = true
}
