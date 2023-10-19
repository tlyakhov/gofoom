package entities

import (
	"image/color"

	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/core"
	"tlyakhov/gofoom/registry"
)

type Player struct {
	AliveEntity `editable:"^"`

	FrameTint color.NRGBA
	Crouching bool
	Inventory []core.AbstractEntity
	Bob       float64
}

func init() {
	registry.Instance().Register(Player{})
}

func NewPlayer(m *core.Map) *Player {
	p := &Player{}
	p.Map = m
	p.Sim = m.Sim
	p.Initialize()
	p.Pos.Original = m.Spawn
	p.Pos.Reset()
	return p
}

func (p *Player) Initialize() {
	p.AliveEntity.Initialize()
	p.Height = constants.PlayerHeight
	p.BoundingRadius = constants.PlayerBoundingRadius
	p.Mass = 70 // kg
}
