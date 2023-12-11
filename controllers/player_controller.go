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

func (pc *PlayerController) Methods() concepts.ControllerMethod {
	return concepts.ControllerAlways
}

func (pc *PlayerController) Target(target *concepts.EntityRef) bool {
	pc.TargetEntity = target
	pc.Player = behaviors.PlayerFromDb(target)
	if pc.Player == nil || !pc.Player.Active {
		return false
	}
	pc.Body = core.BodyFromDb(target)
	if pc.Body == nil || !pc.Body.Active {
		return false
	}
	pc.Alive = behaviors.AliveFromDb(target)
	if pc.Alive == nil || !pc.Alive.Active {
		return false
	}
	return true
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

func MovePlayer(p *core.Body, angle float64) {
	if p.OnGround {
		p.Force[0] += math.Cos(angle*concepts.Deg2rad) * constants.PlayerWalkForce
		p.Force[1] += math.Sin(angle*concepts.Deg2rad) * constants.PlayerWalkForce
	}
}
