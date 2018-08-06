package entity

import (
	"math"

	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/mapping"

	"github.com/tlyakhov/gofoom/constants"
)

type PlayerService struct {
	*mapping.Player
	*AliveEntityService
}

func NewPlayerService(p *mapping.Player) *PlayerService {
	return &PlayerService{Player: p, AliveEntityService: NewAliveEntityService(&p.AliveEntity)}
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
		p.Height = constants.PlayerCrouchHeight
	} else {
		p.Height = constants.PlayerHeight
	}

	if p.Player.HurtTime > 0 {
		//globalGame.frameTint = 255 | ((fast_floor(this.hurtTime * 200 / GAME_CONSTANTS.playerHurtTime) & 0xFF) << 24);
		p.Player.HurtTime--
	}
}

func (p *PlayerService) Hurt(amount float64) {
	p.Hurt(amount)
	p.Player.HurtTime = constants.PlayerHurtTime
}

func (p *PlayerService) Move(angle, lastFrameTime, speed float64) {
	p.Player.Vel.X += math.Cos(angle*concepts.Deg2rad) * constants.PlayerSpeed * speed
	p.Player.Vel.Y += math.Sin(angle*concepts.Deg2rad) * constants.PlayerSpeed * speed
}
