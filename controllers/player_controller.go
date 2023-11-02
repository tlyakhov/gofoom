package controllers

import (
	"image/color"
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
		pc.Body.Height = constants.PlayerCrouchHeight
	} else {
		pc.Body.Height = constants.PlayerHeight
	}

	if pc.Alive.HurtTime > 0 {
		pc.FrameTint = color.NRGBA{0xFF, 0, 0, uint8(pc.Alive.HurtTime * 200 / constants.PlayerHurtTime)}
	}
}

func MovePlayer(p *core.Body, angle float64) {
	if p.OnGround {
		p.Force[0] += math.Cos(angle*concepts.Deg2rad) * constants.PlayerWalkForce
		p.Force[1] += math.Sin(angle*concepts.Deg2rad) * constants.PlayerWalkForce
	}
}
