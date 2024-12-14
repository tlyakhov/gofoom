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
	*core.Body
	Sector *core.Sector
	Player *behaviors.Player

	pos        *concepts.Vector3
	pos2d      *concepts.Vector2
	halfHeight float64
}

func init() {
	ecs.Types().RegisterController(func() ecs.Controller { return &BodyController{} }, 75)
}

func (bc *BodyController) ComponentID() ecs.ComponentID {
	return core.BodyCID
}

func (bc *BodyController) Methods() ecs.ControllerMethod {
	return ecs.ControllerAlways |
		ecs.ControllerRecalculate
}

func (bc *BodyController) EditorPausedMethods() ecs.ControllerMethod {
	return ecs.ControllerRecalculate
}

func (bc *BodyController) Target(target ecs.Attachable, e ecs.Entity) bool {
	bc.Entity = e
	bc.Body = target.(*core.Body)
	if bc.Body == nil || !bc.Body.IsActive() {
		return false
	}
	bc.Player = behaviors.GetPlayer(bc.Body.ECS, bc.Entity)
	if bc.Player != nil && bc.Player.Spawn {
		// If this is a spawn point, skip it
		return false
	}
	bc.Sector = bc.Body.Sector()
	bc.pos = &bc.Pos.Now
	bc.pos2d = bc.pos.To2D()
	bc.halfHeight = bc.Size.Now[1] * 0.5
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

func (bc *BodyController) findBodySector() {
	if bc.Sector != nil && bc.Sector.IsPointInside2D(bc.pos2d) {
		return
	}

	var closestSector *core.Sector

	// This should be optimized
	col := ecs.ColumnFor[core.Sector](bc.ECS, core.SectorCID)
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
		// Find the closest segment and use its sector
		p := bc.Pos.Now.To2D()
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
	// log.Printf("E: %v, P: %v, F: %v, C:%v\n", bc.Entity, bc.pos, floorZ, ceilZ)
	if bc.pos[2]-bc.halfHeight < floorZ {
		bc.pos[2] = floorZ + bc.halfHeight
	}
	if bc.pos[2]+bc.halfHeight > ceilZ {
		bc.pos[2] = ceilZ - bc.halfHeight
	}
	if bc.Sector != closestSector {
		if bc.Sector != nil {
			bc.Exit()
		}
		bc.Enter(closestSector)
	}
}

func (bc *BodyController) Enter(sector *core.Sector) {
	if sector == nil {
		log.Printf("%v tried to enter nil sector", bc.Entity)
		return
	}
	bc.Sector = sector
	bc.Sector.Bodies[bc.Entity] = bc.Body
	bc.Body.SectorEntity = sector.Entity

	if core.GetLight(bc.ECS, bc.Entity) != nil {
		bc.Sector.PVL = append(bc.Sector.PVL, bc.Body)
	}

	if m := core.GetMobile(bc.ECS, bc.Entity); m != nil {
		if m.CrBody != core.CollideNone || m.CrPlayer != core.CollideNone {
			bc.Sector.Colliders[m.Entity] = m
		}
	}

	if bc.Body.OnGround {
		floorZ := bc.Sector.Bottom.ZAt(dynamic.DynamicNow, bc.Pos.Now.To2D())
		p := &bc.Pos.Now
		h := bc.Size.Now[1] * 0.5
		if bc.Sector.Bottom.Target == 0 && p[2]-h < floorZ {
			p[2] = floorZ + h
		}
	}
	BodySectorScript(bc.Sector.EnterScripts, bc.Body, bc.Sector)
}

func (bc *BodyController) Exit() {
	if bc.Sector == nil {
		log.Printf("%v tried to exit nil sector", bc.Entity)
		return
	}
	BodySectorScript(bc.Sector.ExitScripts, bc.Body, bc.Sector)
	delete(bc.Sector.Bodies, bc.Entity)
	// Don't delete out of the PVL, avoid flickering
	if m := core.GetMobile(bc.ECS, bc.Entity); m != nil {
		if m.CrBody != core.CollideNone || m.CrPlayer != core.CollideNone {
			delete(bc.Sector.Colliders, m.Entity)
		}
	}
	bc.SectorEntity = 0
}
