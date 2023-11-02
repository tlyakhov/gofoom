package controllers

import (
	"log"
	"math"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
)

func (bc *BodyController) Enter() {
	if bc.SourceEntity.Nil() {
		log.Printf("%v tried to enter nil sector", bc.TargetEntity.Entity)
		return
	}
	bc.Sector = core.SectorFromDb(bc.SourceEntity)
	bc.Sector.Bodies[bc.TargetEntity.Entity] = *bc.TargetEntity
	bc.Body.SectorEntityRef = bc.SourceEntity

	if bc.Body.OnGround {
		floorZ, _ := bc.Sector.SlopedZNow(bc.Body.Pos.Now.To2D())
		p := &bc.Body.Pos.Now
		if bc.Sector.FloorTarget.Nil() && p[2] < floorZ {
			p[2] = floorZ
		}
	}
}

func (bc *BodyController) Exit() {
	if bc.Sector == nil {
		log.Printf("%v tried to exit nil sector", bc.TargetEntity.Entity)
		return
	}
	delete(bc.Sector.Bodies, bc.Body.Entity)
	bc.Body.SectorEntityRef = nil
}

func (bc *BodyController) Containment() {
	f := &bc.Body.Force
	if bc.Body.Mass > 0 {
		// Weight = g*m
		f[2] -= constants.Gravity * bc.Body.Mass
		v := &bc.Body.Vel.Now
		if !v.Zero() {
			// Air drag
			r := bc.Body.BoundingRadius * constants.MetersPerUnit
			crossSectionArea := math.Pi * r * r
			drag := concepts.Vector3{v[0], v[1], v[2]}
			drag.MulSelf(drag.Length())
			drag.MulSelf(-0.5 * constants.AirDensity * crossSectionArea * constants.SphereDragCoefficient)
			f.AddSelf(&drag)
			if bc.Body.OnGround {
				// Kinetic friction
				drag.From(v)
				g := concepts.Vector3{0, 0, constants.Gravity * bc.Body.Mass}
				drag.MulSelf(-bc.Sector.FloorFriction * bc.Sector.FloorNormal.Dot(&g))
				f.AddSelf(&drag)
			}
			//log.Printf("%v\n", drag)
		}
	}

	set := bc.NewControllerSet()

	if !bc.Sector.FloorMaterial.Nil() && bc.Body.Pos.Now[2] <= bc.Sector.BottomZ.Now {
		set.Act(bc.TargetEntity, bc.Sector.FloorMaterial, concepts.ControllerContact)
	}
	if !bc.Sector.CeilMaterial.Nil() && bc.Body.Pos.Now[2] >= bc.Sector.TopZ.Now {
		set.Act(bc.TargetEntity, bc.Sector.CeilMaterial, concepts.ControllerContact)
	}

	bodyTop := bc.Body.Pos.Now[2] + bc.Body.Height
	floorZ, ceilZ := bc.Sector.SlopedZNow(bc.Body.Pos.Now.To2D())

	bc.Body.OnGround = false
	if !bc.Sector.FloorTarget.Nil() && bodyTop < floorZ {
		set.Act(bc.TargetEntity, nil, concepts.ControllerExit)
		set.Act(bc.TargetEntity, bc.Sector.FloorTarget, concepts.ControllerEnter)
		bc.Sector = core.SectorFromDb(bc.Sector.FloorTarget)
		_, ceilZ = bc.Sector.SlopedZNow(bc.Body.Pos.Now.To2D())
		bc.Body.Pos.Now[2] = ceilZ - bc.Body.Height - 1.0
	} else if !bc.Sector.FloorTarget.Nil() && bc.Body.Pos.Now[2] <= floorZ && bc.Body.Vel.Now[2] > 0 {
		bc.Body.Vel.Now[2] = constants.PlayerJumpForce
	} else if bc.Sector.FloorTarget.Nil() && bc.Body.Pos.Now[2] <= floorZ {
		dist := bc.Sector.FloorNormal[2] * (floorZ - bc.Body.Pos.Now[2])
		delta := bc.Sector.FloorNormal.Mul(dist)
		bc.Body.Vel.Now.AddSelf(delta)
		bc.Body.Pos.Now.AddSelf(delta)
		bc.Body.OnGround = true
	}

	if !bc.Sector.CeilTarget.Nil() && bodyTop > ceilZ {
		set.Act(bc.TargetEntity, nil, concepts.ControllerExit)
		set.Act(bc.TargetEntity, bc.Sector.CeilTarget, concepts.ControllerEnter)
		bc.Sector = core.SectorFromDb(bc.Sector.CeilTarget)
		floorZ, _ = bc.Sector.SlopedZNow(bc.Body.Pos.Now.To2D())
		bc.Body.Pos.Now[2] = floorZ - bc.Body.Height + 1.0
	} else if bc.Sector.CeilTarget.Nil() && bodyTop >= ceilZ {
		dist := -bc.Sector.CeilNormal[2] * (bodyTop - ceilZ + 1.0)
		delta := bc.Sector.CeilNormal.Mul(dist)
		bc.Body.Vel.Now.AddSelf(delta)
		bc.Body.Pos.Now.AddSelf(delta)
	}
}
