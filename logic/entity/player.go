package entity

import (
	"image/color"
	"math"

	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/entities"

	"github.com/tlyakhov/gofoom/constants"
)

type PlayerService struct {
	*entities.Player
	*AliveEntityService
}

func NewPlayerService(p *entities.Player) *PlayerService {
	return &PlayerService{Player: p, AliveEntityService: NewAliveEntityService(&p.AliveEntity, p)}
}

func (p *PlayerService) Frame(lastFrameTime float64) {
	p.AliveEntityService.PhysicalEntityService.Frame(lastFrameTime)
	if p.Player.Sector == nil {
		return
	}

	if p.Player.Vel.Z <= 0 && p.Player.Vel.Z >= -0.001 {
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
	p.Player.Vel.X += math.Cos(angle*concepts.Deg2rad) * constants.PlayerSpeed * speed
	p.Player.Vel.Y += math.Sin(angle*concepts.Deg2rad) * constants.PlayerSpeed * speed
}
