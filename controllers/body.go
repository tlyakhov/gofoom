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

type BodyController struct {
	concepts.BaseController
	Body   *core.Body
	Sector *core.Sector
	Player *behaviors.Player

	collidedSegments []*core.SectorSegment
	pos              *concepts.Vector3
	pos2d            *concepts.Vector2
	halfHeight       float64
}

func init() {
	concepts.DbTypes().RegisterController(&BodyController{}, 80)
}

func (bc *BodyController) ComponentIndex() int {
	return core.BodyComponentIndex
}

func (bc *BodyController) Methods() concepts.ControllerMethod {
	return concepts.ControllerAlways |
		concepts.ControllerRecalculate |
		concepts.ControllerLoaded
}

func (bc *BodyController) Target(target concepts.Attachable) bool {
	bc.Body = target.(*core.Body)
	if !bc.Body.IsActive() {
		return false
	}
	bc.Player = behaviors.PlayerFromDb(bc.Body.DB, bc.Body.Entity)
	if bc.Player != nil && bc.Player.Spawn {
		// If this is a spawn point, skip it
		return false
	}
	bc.Sector = bc.Body.Sector()
	bc.pos = &bc.Body.Pos.Now
	bc.pos2d = bc.pos.To2D()
	bc.halfHeight = bc.Body.Size.Now[1] * 0.5
	return true
}

func (bc *BodyController) ResetForce() {
	bc.Body.Force[2] = 0.0
	bc.Body.Force[1] = 0.0
	bc.Body.Force[0] = 0.0
}

func (bc *BodyController) Forces() {
	f := &bc.Body.Force
	if bc.Body.Mass > 0 {
		// Weight = g*m
		g := bc.Sector.Gravity
		g.MulSelf(bc.Body.Mass)
		f.AddSelf(&g)
		v := &bc.Body.Vel.Now
		if !v.Zero() {
			// Air drag
			r := bc.Body.Size.Now[0] * 0.5 * constants.MetersPerUnit
			crossSectionArea := math.Pi * r * r
			drag := concepts.Vector3{v[0], v[1], v[2]}
			drag.MulSelf(drag.Length())
			drag.MulSelf(-0.5 * constants.AirDensity * crossSectionArea * constants.SphereDragCoefficient)
			f.AddSelf(&drag)
			if bc.Body.OnGround {
				// Kinetic friction
				drag.From(v)
				drag.MulSelf(-bc.Sector.FloorFriction * bc.Sector.FloorNormal.Dot(g.MulSelf(-1)))
				f.AddSelf(&drag)
			}
			//log.Printf("%v\n", drag)
		}
	}
}

func (bc *BodyController) Always() {
	if bc.Body.Mass == 0 {
		// Reset force for next frame
		bc.ResetForce()
		return
	}
	if bc.Sector != nil {
		bc.Forces()
	}
	// Our physics are impulse-based. We do semi-implicit Euler calculations
	// at each time step, and apply constraints (e.g. collision) directly to the velocities
	// f = ma
	// a = f/m
	// v = ∫a dt
	// p = ∫v dt
	bc.Body.Vel.Now.AddSelf(bc.Body.Force.Mul(constants.TimeStepS / bc.Body.Mass))
	if bc.Body.Vel.Now.Length2() > constants.VelocityEpsilon {
		speed := bc.Body.Vel.Now.Length() * constants.TimeStepS
		steps := concepts.Min(concepts.Max(int(speed/constants.CollisionCheck), 1), 10)
		dt := constants.TimeStepS / float64(steps)
		for step := 0; step < steps; step++ {
			bc.Body.Pos.Now.AddSelf(bc.Body.Vel.Now.Mul(dt * constants.UnitsPerMeter))
			// Constraint impulses
			bc.Collide()
		}
	}
	// Reset force for next frame
	bc.ResetForce()
}

func (bc *BodyController) Recalculate() {
	bc.Collide()
}

func (bc *BodyController) Loaded() {
	bc.Collide()
}
