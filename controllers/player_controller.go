package controllers

import (
	"math"

	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"

	"tlyakhov/gofoom/constants"
)

type PlayerController struct {
	concepts.BaseController
	*behaviors.Player
	Alive *behaviors.Alive
	Body  *core.Body
}

func init() {
	concepts.DbTypes().RegisterController(&PlayerController{})
}

func (pc *PlayerController) ComponentIndex() int {
	return behaviors.PlayerComponentIndex
}

func (pc *PlayerController) Methods() concepts.ControllerMethod {
	return concepts.ControllerAlways
}

func (pc *PlayerController) Target(target concepts.Attachable) bool {
	pc.Player = target.(*behaviors.Player)
	if !pc.Player.IsActive() {
		return false
	}
	pc.Body = core.BodyFromDb(pc.Player.EntityRef)
	if !pc.Body.IsActive() {
		return false
	}
	pc.Alive = behaviors.AliveFromDb(pc.Player.EntityRef)
	return pc.Alive.IsActive()
}

func (pc *PlayerController) Always() {
	pc.Bob += pc.Body.Vel.Now.To2D().Length() / 100.0
	for pc.Bob > math.Pi*2 {
		pc.Bob -= math.Pi * 2
	}

	if pc.Crouching {
		pc.Body.Size.Now[2] = constants.PlayerCrouchHeight
	} else {
		pc.Body.Size.Now[2] = constants.PlayerHeight
	}

	allCooldowns := 0.0
	maxCooldown := 0.0
	for _, d := range pc.Alive.Damages {
		allCooldowns += d.Cooldown.Render
		maxCooldown += d.Cooldown.Original
	}

	if pc.Underwater() {
		pc.FrameTint = concepts.Vector4{75.0 / 255.0, 147.0 / 255.0, 1, 90.0 / 255.0}
	} else {
		pc.FrameTint = concepts.Vector4{}
	}

	if allCooldowns > 0 && maxCooldown > 0 {
		a := allCooldowns * 0.6 / maxCooldown
		pc.FrameTint.MulSelf(1.0 - a)
		pc.FrameTint.AddSelf(&concepts.Vector4{1, 0, 0, a})
	}
}

func MovePlayer(p *core.Body, angle float64, direct bool) {
	dy, dx := math.Sincos(angle * concepts.Deg2rad)
	dy *= constants.PlayerWalkForce
	dx *= constants.PlayerWalkForce

	if direct {
		p.Pos.Now[0] += dx * 0.1 / p.Mass
		p.Pos.Now[1] += dy * 0.1 / p.Mass
	} else {
		if p.OnGround {
			p.Force[0] += dx
			p.Force[1] += dy
		}
	}
}
