package entity

import (
	"image/color"
	"math"

	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/core"
	"tlyakhov/gofoom/entities"

	"tlyakhov/gofoom/constants"
)

type PlayerService struct {
	*entities.Player
	*AliveEntityService
}

func NewPlayerService(p *entities.Player) *PlayerService {
	return &PlayerService{Player: p, AliveEntityService: NewAliveEntityService(&p.AliveEntity, p)}
}

func (p *PlayerService) Frame(sim *core.Simulation) {
	p.Bob += p.Player.Vel.Now.Length() / 6.0
	for p.Bob > math.Pi*2 {
		p.Bob -= math.Pi * 2
	}
	p.AliveEntityService.PhysicalEntityService.Frame(sim)
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
	p.AliveEntityService.Hurt(amount)
	p.Player.HurtTime = constants.PlayerHurtTime
}

func (p *PlayerService) Move(angle, lastFrameTime float64) {
	p.Player.Vel.Now[0] += math.Cos(angle*concepts.Deg2rad) * constants.PlayerSpeed * constants.TimeStep
	p.Player.Vel.Now[1] += math.Sin(angle*concepts.Deg2rad) * constants.PlayerSpeed * constants.TimeStep
}
