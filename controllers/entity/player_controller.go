package entity

import (
	"image/color"
	"math"

	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/entities"

	"tlyakhov/gofoom/constants"
)

type PlayerService struct {
	*entities.Player
	*AliveEntityController
}

func NewPlayerController(p *entities.Player) *PlayerService {
	return &PlayerService{Player: p, AliveEntityController: NewAliveEntityController(&p.AliveEntity, p)}
}

func (p *PlayerService) Frame() {
	p.Bob += p.Player.Vel.Now.Length() / 6.0
	for p.Bob > math.Pi*2 {
		p.Bob -= math.Pi * 2
	}
	p.AliveEntityController.PhysicalEntityController.Frame()
	if p.Player.Sector == nil {
		return
	}

	if p.Crouching {
		p.Player.Height = constants.PlayerCrouchHeight
	} else {
		p.Player.Height = constants.PlayerHeight
	}

	if p.Player.HurtTime > 0 {
		p.FrameTint = color.NRGBA{0xFF, 0, 0, uint8(p.Player.HurtTime * 200 / constants.PlayerHurtTime)}
		p.Player.HurtTime--
	}
}

func (p *PlayerService) Hurt(amount float64) {
	p.AliveEntityController.Hurt(amount)
	p.Player.HurtTime = constants.PlayerHurtTime
}

func (p *PlayerService) Move(angle float64) {
	p.Player.Vel.Now[0] += math.Cos(angle*concepts.Deg2rad) * constants.PlayerSpeed * constants.TimeStep
	p.Player.Vel.Now[1] += math.Sin(angle*concepts.Deg2rad) * constants.PlayerSpeed * constants.TimeStep
}
