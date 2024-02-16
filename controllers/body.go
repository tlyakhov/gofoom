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
	concepts.DbTypes().RegisterController(&BodyController{})
}

func (bc *BodyController) ComponentIndex() int {
	return core.BodyComponentIndex
}

func (bc *BodyController) Priority() int {
	return 80
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
	bc.Sector = bc.Body.Sector()
	bc.Player = behaviors.PlayerFromDb(bc.Body.EntityRef)
	bc.pos = &bc.Body.Pos.Now
	bc.pos2d = bc.pos.To2D()
	bc.halfHeight = bc.Body.Size.Now[1] * 0.5
	return true
}

func (bc *BodyController) RemoveBody() {
	// TODO: poorly implemented
	if bc.Sector != nil {
		delete(bc.Sector.Bodies, bc.Body.Entity)
		bc.Sector = nil
		bc.Body.SectorEntityRef.Now = nil
		//return
	}
	panic("BodyController.RemoveBody is broken")
}

func (bc *BodyController) ResetForce() {
	bc.Body.Force[2] = 0.0
	bc.Body.Force[1] = 0.0
	bc.Body.Force[0] = 0.0
}

func (bc *BodyController) Physics() {
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

	halfHeight := bc.Body.Size.Now[1] * 0.5
	bodyTop := bc.Body.Pos.Now[2] + halfHeight
	floorZ, ceilZ := bc.Sector.SlopedZNow(bc.Body.Pos.Now.To2D())

	bc.Body.OnGround = false
	if !bc.Sector.FloorTarget.Nil() && bodyTop < floorZ {
		bc.Exit()
		bc.Enter(bc.Sector.FloorTarget)
		_, ceilZ = bc.Sector.SlopedZNow(bc.Body.Pos.Now.To2D())
		bc.Body.Pos.Now[2] = ceilZ - halfHeight - 1.0
	} else if !bc.Sector.FloorTarget.Nil() && bc.Body.Pos.Now[2]-halfHeight <= floorZ && bc.Body.Vel.Now[2] > 0 {
		bc.Body.Vel.Now[2] = constants.PlayerJumpForce
	} else if bc.Sector.FloorTarget.Nil() && bc.Body.Pos.Now[2]-halfHeight <= floorZ {
		dist := bc.Sector.FloorNormal[2] * (floorZ - (bc.Body.Pos.Now[2] - halfHeight))
		delta := bc.Sector.FloorNormal.Mul(dist)
		bc.Body.Vel.Now.AddSelf(delta)
		bc.Body.Pos.Now.AddSelf(delta)
		bc.Body.OnGround = true
		BodySectorScript(bc.Sector.FloorScripts, bc.Body.EntityRef, bc.Sector.EntityRef)
	}

	if !bc.Sector.CeilTarget.Nil() && bodyTop > ceilZ {
		bc.Exit()
		bc.Enter(bc.Sector.CeilTarget)
		bc.Sector = core.SectorFromDb(bc.Sector.CeilTarget)
		floorZ, _ = bc.Sector.SlopedZNow(bc.Body.Pos.Now.To2D())
		bc.Body.Pos.Now[2] = floorZ + halfHeight + 1.0
	} else if bc.Sector.CeilTarget.Nil() && bodyTop >= ceilZ {
		dist := -bc.Sector.CeilNormal[2] * (bodyTop - ceilZ + 1.0)
		delta := bc.Sector.CeilNormal.Mul(dist)
		bc.Body.Vel.Now.AddSelf(delta)
		bc.Body.Pos.Now.AddSelf(delta)
		BodySectorScript(bc.Sector.CeilScripts, bc.Body.EntityRef, bc.Sector.EntityRef)
	}
}

func (bc *BodyController) Always() {
	if bc.Body.Mass == 0 {
		// Reset force for next frame
		bc.ResetForce()
		return
	}
	if bc.Sector != nil {
		bc.Physics()
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
