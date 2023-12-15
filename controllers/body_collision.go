package controllers

import (
	"log"
	"math"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
)

func BodySectorScript(scripts []core.Script, ibody, isector *concepts.EntityRef) {
	for _, script := range scripts {
		script.Vars["body"] = ibody
		script.Vars["sector"] = isector
		script.Act()
	}
}

func (bc *BodyController) Enter(sectorRef *concepts.EntityRef) {
	if sectorRef.Nil() {
		log.Printf("%v tried to enter nil sector", bc.TargetEntity.Entity)
		return
	}
	sector := core.SectorFromDb(sectorRef)
	if sector == nil {
		log.Printf("%v tried to enter entity %v that's not a sector", bc.TargetEntity.Entity, sectorRef.String())
		return
	}
	bc.Sector = sector
	bc.Sector.Bodies[bc.TargetEntity.Entity] = bc.TargetEntity
	bc.Body.SectorEntityRef = sectorRef

	if bc.Body.OnGround {
		floorZ, _ := bc.Sector.SlopedZNow(bc.Body.Pos.Now.To2D())
		p := &bc.Body.Pos.Now
		if bc.Sector.FloorTarget.Nil() && p[2] < floorZ {
			p[2] = floorZ
		}
	}
	BodySectorScript(bc.Sector.EnterScripts, bc.TargetEntity, bc.Sector.Ref())
}

func (bc *BodyController) Exit() {
	if bc.Sector == nil {
		log.Printf("%v tried to exit nil sector", bc.TargetEntity.Entity)
		return
	}
	BodySectorScript(bc.Sector.ExitScripts, bc.TargetEntity, bc.Sector.Ref())
	delete(bc.Sector.Bodies, bc.Body.Entity)
	bc.Body.SectorEntityRef = nil
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
			r := bc.Body.BoundingRadius * constants.MetersPerUnit
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

	bodyTop := bc.Body.Pos.Now[2] + bc.Body.Size.Now[2]
	floorZ, ceilZ := bc.Sector.SlopedZNow(bc.Body.Pos.Now.To2D())

	bc.Body.OnGround = false
	if !bc.Sector.FloorTarget.Nil() && bodyTop < floorZ {
		bc.Exit()
		bc.Enter(bc.Sector.FloorTarget)
		_, ceilZ = bc.Sector.SlopedZNow(bc.Body.Pos.Now.To2D())
		bc.Body.Pos.Now[2] = ceilZ - bc.Body.Size.Now[2] - 1.0
	} else if !bc.Sector.FloorTarget.Nil() && bc.Body.Pos.Now[2] <= floorZ && bc.Body.Vel.Now[2] > 0 {
		bc.Body.Vel.Now[2] = constants.PlayerJumpForce
	} else if bc.Sector.FloorTarget.Nil() && bc.Body.Pos.Now[2] <= floorZ {
		dist := bc.Sector.FloorNormal[2] * (floorZ - bc.Body.Pos.Now[2])
		delta := bc.Sector.FloorNormal.Mul(dist)
		bc.Body.Vel.Now.AddSelf(delta)
		bc.Body.Pos.Now.AddSelf(delta)
		bc.Body.OnGround = true
		BodySectorScript(bc.Sector.FloorScripts, bc.TargetEntity, bc.Sector.Ref())
	}

	if !bc.Sector.CeilTarget.Nil() && bodyTop > ceilZ {
		bc.Exit()
		bc.Enter(bc.Sector.CeilTarget)
		bc.Sector = core.SectorFromDb(bc.Sector.CeilTarget)
		floorZ, _ = bc.Sector.SlopedZNow(bc.Body.Pos.Now.To2D())
		bc.Body.Pos.Now[2] = floorZ - bc.Body.Size.Now[2] + 1.0
	} else if bc.Sector.CeilTarget.Nil() && bodyTop >= ceilZ {
		dist := -bc.Sector.CeilNormal[2] * (bodyTop - ceilZ + 1.0)
		delta := bc.Sector.CeilNormal.Mul(dist)
		bc.Body.Vel.Now.AddSelf(delta)
		bc.Body.Pos.Now.AddSelf(delta)
		BodySectorScript(bc.Sector.CeilScripts, bc.TargetEntity, bc.Sector.Ref())
	}
}
