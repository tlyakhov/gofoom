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
	*AliveEntityService
}

func NewPlayerService(p *entities.Player) *PlayerService {
	return &PlayerService{Player: p, AliveEntityService: NewAliveEntityService(&p.AliveEntity, p)}
}

func (p *PlayerService) Frame(lastFrameTime float64) {
	p.Bob += p.Player.Vel.Length() / 6.0
	for p.Bob > math.Pi*2 {
		p.Bob -= math.Pi * 2
	}
	p.AliveEntityService.PhysicalEntityService.Frame(lastFrameTime)
	if p.Player.Sector == nil {
		return
	}

	if p.Player.Vel[2] <= 0 && p.Player.Vel[2] >= -0.001 {
		p.Standing = true
	} else {
		p.Standing = false
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

func (p *PlayerService) Move(angle, lastFrameTime, speed float64) {
	p.Player.Vel[0] += math.Cos(angle*concepts.Deg2rad) * constants.PlayerSpeed * speed
	p.Player.Vel[1] += math.Sin(angle*concepts.Deg2rad) * constants.PlayerSpeed * speed
}
