// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"log"
	"math"
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/dynamic"
	"tlyakhov/gofoom/ecs"
)

type BodyController struct {
	ecs.BaseController
	Body   *core.Body
	Sector *core.Sector
	Player *behaviors.Player

	pos        *concepts.Vector3
	pos2d      *concepts.Vector2
	halfHeight float64
}

func init() {
	ecs.Types().RegisterController(&BodyController{}, 75)
}

func (bc *BodyController) ComponentID() ecs.ComponentID {
	return core.BodyCID
}

func (bc *BodyController) Methods() ecs.ControllerMethod {
	return ecs.ControllerAlways |
		ecs.ControllerRecalculate |
		ecs.ControllerLoaded
}

func (bc *BodyController) Target(target ecs.Attachable) bool {
	bc.Body = target.(*core.Body)
	if !bc.Body.IsActive() {
		return false
	}
	bc.Player = behaviors.GetPlayer(bc.Body.ECS, bc.Body.Entity)
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

func (bc *BodyController) Always() {
	//if bc.Sector == nil {
	// Try to put this body into a sector
	//	bc.Collide()
	//	bc.findBodySector()
	//}
}

func (bc *BodyController) Recalculate() {
	//bc.Collide()
	bc.findBodySector()
}

func (bc *BodyController) Loaded() {
	//bc.Collide()
	bc.findBodySector()
}

func (bc *BodyController) findBodySector() {
	var closestSector *core.Sector

	col := ecs.ColumnFor[core.Sector](bc.Body.ECS, core.SectorCID)
	for i := range col.Cap() {
		sector := col.Value(i)
		if sector == nil {
			continue
		}
		if sector.IsPointInside2D(bc.pos2d) {
			closestSector = sector
			break
		}
	}

	if closestSector == nil {
		p := bc.Body.Pos.Now.To2D()
		var closestSeg *core.SectorSegment
		closestDistance2 := math.MaxFloat64
		for i := range col.Cap() {
			sector := col.Value(i)
			if sector == nil {
				continue
			}
			for _, seg := range sector.Segments {
				dist2 := seg.DistanceToPoint2(p)
				if closestSector == nil || dist2 < closestDistance2 {
					closestDistance2 = dist2
					closestSector = sector
					closestSeg = seg
				}
			}
		}
		p = closestSeg.ClosestToPoint(p)
		bc.Body.Pos.Input[0] = p[0]
		bc.Body.Pos.Input[1] = p[1]
	}

	floorZ, ceilZ := closestSector.ZAt(dynamic.DynamicNow, bc.pos2d)
	//log.Printf("F: %v, C:%v\n", floorZ, ceilZ)
	if bc.pos[2]-bc.halfHeight < floorZ || bc.pos[2]+bc.halfHeight > ceilZ {
		//log.Printf("Moved body %v to closest sector and adjusted Z from %v to %v", mc.Body.Entity, p[2], floorZ)
		bc.pos[2] = floorZ + bc.halfHeight
	}
	bc.Enter(closestSector.Entity)
	// Don't mark as collided because this is probably an initialization.
}

func (bc *BodyController) Enter(eSector ecs.Entity) {
	if eSector == 0 {
		log.Printf("%v tried to enter nil sector", bc.Body.Entity)
		return
	}
	sector := core.GetSector(bc.Body.ECS, eSector)
	if sector == nil {
		log.Printf("%v tried to enter entity %v that's not a sector", bc.Body.Entity, eSector.Format(bc.Body.ECS))
		return
	}
	bc.Sector = sector
	bc.Sector.Bodies[bc.Body.Entity] = bc.Body
	bc.Body.SectorEntity = eSector

	if bc.Body.OnGround {
		floorZ := bc.Sector.Bottom.ZAt(dynamic.DynamicNow, bc.Body.Pos.Now.To2D())
		p := &bc.Body.Pos.Now
		h := bc.Body.Size.Now[1] * 0.5
		if bc.Sector.Bottom.Target == 0 && p[2]-h < floorZ {
			p[2] = floorZ + h
		}
	}
	BodySectorScript(bc.Sector.EnterScripts, bc.Body, bc.Sector)
}

func (bc *BodyController) Exit() {
	if bc.Sector == nil {
		log.Printf("%v tried to exit nil sector", bc.Body.Entity)
		return
	}
	BodySectorScript(bc.Sector.ExitScripts, bc.Body, bc.Sector)
	delete(bc.Sector.Bodies, bc.Body.Entity)
	bc.Body.SectorEntity = 0
}
