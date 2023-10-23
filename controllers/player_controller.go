package controllers

import (
	"image/color"
	"math"

	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"

	"tlyakhov/gofoom/constants"
)

func FramePlayer(mob *core.Mob, p *behaviors.Player, a *behaviors.Alive) {
	p.Bob += mob.Vel.Now.Length() / 100.0
	for p.Bob > math.Pi*2 {
		p.Bob -= math.Pi * 2
	}

	if p.Crouching {
		mob.Height = constants.PlayerCrouchHeight
	} else {
		mob.Height = constants.PlayerHeight
	}

	if a.HurtTime > 0 {
		p.FrameTint = color.NRGBA{0xFF, 0, 0, uint8(a.HurtTime * 200 / constants.PlayerHurtTime)}
		a.HurtTime--
	}
}

func MovePlayer(p *core.Mob, angle float64) {
	if p.OnGround {
		p.Force[0] += math.Cos(angle*concepts.Deg2rad) * constants.PlayerWalkForce
		p.Force[1] += math.Sin(angle*concepts.Deg2rad) * constants.PlayerWalkForce
	}
}
