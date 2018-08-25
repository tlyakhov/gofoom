package entities

import (
	"image/color"

	"github.com/tlyakhov/gofoom/constants"
	"github.com/tlyakhov/gofoom/core"
	"github.com/tlyakhov/gofoom/registry"
)

type Player struct {
	AliveEntity `editable:"^"`

	FrameTint color.NRGBA
	Standing  bool
	Crouching bool
	Inventory []core.AbstractEntity
}

func init() {
	registry.Instance().Register(Player{})
}

func NewPlayer(m *core.Map) *Player {
	p := &Player{}
	p.Map = m
	p.Initialize()
	p.Pos = m.Spawn
	return p
}

func (p *Player) Initialize() {
	p.AliveEntity.Initialize()
	p.Height = constants.PlayerHeight
	p.BoundingRadius = constants.PlayerBoundingRadius
	p.Standing = true
	p.Weight = 1
}
