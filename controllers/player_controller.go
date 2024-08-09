// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

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
	pc.Body = core.BodyFromDb(pc.Player.DB, pc.Player.Entity)
	if !pc.Body.IsActive() {
		return false
	}
	pc.Alive = behaviors.AliveFromDb(pc.Player.DB, pc.Player.Entity)
	return pc.Alive.IsActive()
}

func (pc *PlayerController) Always() {
	uw := pc.Underwater()

	if uw {
		pc.Bob += 0.015
	} else {
		pc.Bob += pc.Body.Vel.Now.To2D().Length() / 64.0
	}

	for pc.Bob > math.Pi*2 {
		pc.Bob -= math.Pi * 2
	}

	// TODO: There's a bug here: this can cause a player<->floor collision that
	// has to be resolved by shoving the player upwards, making an uncrouch into
	// an unintentional jump.
	if pc.Crouching {
		pc.Body.Size.Now[1] = constants.PlayerCrouchHeight
	} else {
		pc.Body.Size.Now[1] = constants.PlayerHeight
	}

	bob := math.Sin(pc.Bob) * 1.5
	pc.CameraZ = pc.Body.Pos.Render[2] + pc.Body.Size.Render[1]*0.5 + bob

	if sector := pc.Body.Sector(); sector != nil {
		fz, cz := sector.PointZ(concepts.DynamicRender, pc.Body.Pos.Render.To2D())
		fz += constants.IntersectEpsilon
		cz -= constants.IntersectEpsilon
		if pc.CameraZ < fz {
			pc.CameraZ = fz
		}
		if pc.CameraZ > cz {
			pc.CameraZ = cz
		}
	}

	if uw {
		pc.FrameTint = concepts.Vector4{0.29, 0.58, 1, 0.35}
	} else {
		pc.FrameTint = concepts.Vector4{}
	}

	pc.Alive.Tint(&pc.FrameTint)
}

func MovePlayer(p *core.Body, angle float64, direct bool) {
	uw := behaviors.UnderwaterFromDb(p.DB, p.SectorEntity) != nil
	dy, dx := math.Sincos(angle * concepts.Deg2rad)
	dy *= constants.PlayerWalkForce
	dx *= constants.PlayerWalkForce
	if direct {
		p.Pos.Now[0] += dx * 0.1 / p.Mass
		p.Pos.Now[1] += dy * 0.1 / p.Mass
	} else {
		if uw || p.OnGround {
			p.Force[0] += dx
			p.Force[1] += dy
		}
	}
}
