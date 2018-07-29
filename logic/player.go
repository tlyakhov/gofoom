package logic

import (
	"math"

	"github.com/tlyakhov/gofoom/constants"
)

func (p *Player) Frame(lastFrameTime float64) {
	p.Entity.Frame(lastFrameTime)

	if p.Sector == nil {
		return
	}

	if p.Vel.Z <= 0 && p.Vel.Z >= -0.001 {
		p.Standing = true
	} else {
		p.Standing = false
	}

	if p.Crouching {
		p.Height = constants.PlayerCrouchHeight
	} else {
		p.Height = constants.PlayerHeight
	}

	if p.HurtTime > 0 {
		//globalGame.frameTint = 255 | ((fast_floor(this.hurtTime * 200 / GAME_CONSTANTS.playerHurtTime) & 0xFF) << 24);
		p.HurtTime--
	}
}

func (p *Player) Hurt(amount float64) {
	p.Entity.Hurt(amount)
	p.HurtTime = constants.PlayerHurtTime
}

func (p *Player) Move(angle, lastFrameTime, speed float64) {
	p.Vel.X += math.Cos(angle*deg2rad) * constants.PlayerSpeed * speed
	p.Vel.Y += math.Sin(angle*deg2rad) * constants.PlayerSpeed * speed
}
