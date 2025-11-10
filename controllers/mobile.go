// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"math"
	"math/rand"

	"tlyakhov/gofoom/components/audio"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/ecs"
)

type MobileController struct {
	BodyController
	*core.Mobile

	collidedSegments []*core.Segment
}

func init() {
	ecs.Types().RegisterController(func() ecs.Controller { return &MobileController{} }, 80)
}

func (mc *MobileController) ComponentID() ecs.ComponentID {
	return core.MobileCID
}

func (mc *MobileController) Methods() ecs.ControllerMethod {
	return ecs.ControllerFrame |
		ecs.ControllerRecalculate
}

func (mc *MobileController) EditorPausedMethods() ecs.ControllerMethod {
	return ecs.ControllerRecalculate
}

func (mc *MobileController) Target(target ecs.Component, e ecs.Entity) bool {
	mc.BaseController.Entity = e
	mc.Mobile = target.(*core.Mobile)
	if !mc.Mobile.IsActive() {
		return false
	}
	return mc.BodyController.Target(core.GetBody(mc.BaseController.Entity), e)
}

func (mc *MobileController) ResetForce() {
	mc.Force[2] = 0.0
	mc.Force[1] = 0.0
	mc.Force[0] = 0.0
}

func (mc *MobileController) Forces() {
	f := &mc.Force
	if mc.Mass > 0 {
		// Weight = g*m
		g := mc.Sector.Gravity
		g.MulSelf(mc.Mass)
		f.AddSelf(&g)
		v := &mc.Vel.Now
		if !v.Zero() {
			// Air drag
			r := mc.Body.Size.Now[0] * 0.5 * constants.MetersPerUnit
			crossSectionArea := math.Pi * r * r
			drag := concepts.Vector3{v[0], v[1], v[2]}
			drag.MulSelf(drag.Length())
			drag.MulSelf(-0.5 * constants.AirDensity * crossSectionArea * constants.SphereDragCoefficient)
			f.AddSelf(&drag)
			if mc.Body.OnGround {
				// Kinetic friction
				drag.From(v)
				drag.MulSelf(-mc.Sector.FloorFriction * mc.Sector.Bottom.Normal.Dot(g.MulSelf(-1)))
				f.AddSelf(&drag)
			}
			//log.Printf("%v\n", drag)
		}
	}
}

func (mc *MobileController) Frame() {
	if mc.Mass == 0 {
		// Reset force for next frame
		mc.ResetForce()
		return
	}
	if mc.Sector != nil {
		mc.Forces()
	} else {
		// Try to put this body into a sector
		mc.Collide()
	}
	// Our physics are impulse-based. We do semi-implicit Euler calculations
	// at each time step, and apply constraints (e.g. collision) directly to the velocities
	// f = ma
	// a = f/m
	// v = ∫a dt
	// p = ∫v dt
	prePos := mc.Body.Pos.Now
	mc.Vel.Now.AddSelf(mc.Force.Mul(constants.TimeStepS / mc.Mass))
	speedSquared := mc.Vel.Now.Length2()
	if speedSquared > constants.VelocityEpsilon {
		speed := mc.Vel.Now.Length() * constants.TimeStepS
		steps := min(max(int(speed/constants.CollisionCheck), 1), 10)
		dt := constants.TimeStepS / float64(steps)
		for range steps {
			mc.Body.Pos.Now.AddSelf(mc.Vel.Now.Mul(dt * constants.UnitsPerMeter))
			// Constraint impulses
			mc.Collide()
		}
	}
	if mc.StepSound != 0 {
		// TODO: Handle flying things
		if mc.OnGround {
			// How much did we actually move?
			mc.MovementSoundDistance += mc.Body.Pos.Now.Dist(&prePos)
		}
		// TODO: Parameterize this
		stepDist := 70 + rand.Float64()*10
		if mc.MovementSoundDistance > stepDist || speedSquared <= constants.VelocityEpsilon {
			event, _ := audio.PlaySound(mc.StepSound, mc.Body.Entity, mc.Body.Entity.String()+" step", false)
			if event != nil {
				// TODO: Parameterize this
				event.Offset[2] = -mc.Body.Size.Now[1] * 0.5
				event.Offset[1] = 0
				event.Offset[0] = 0
				event.SetPitchMultiplier(0.9 + rand.Float64()*0.2)
			}
			mc.MovementSoundDistance = 0
		}
	}
	// Reset force for next frame
	mc.ResetForce()

	// Update quadtree
	core.QuadTree.Update(mc.Body)
}

func (mc *MobileController) Recalculate() {
	mc.Collide()
}
