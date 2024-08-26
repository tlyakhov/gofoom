// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"log"
	"math"
	"math/rand/v2"
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/ecs"
)

func BodySectorScript(scripts []*core.Script, body *core.Body, sector *core.Sector) {
	for _, script := range scripts {
		script.Vars["body"] = body
		script.Vars["sector"] = sector
		script.Act()
	}
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
		floorZ := bc.Sector.Bottom.ZAt(ecs.DynamicNow, bc.Body.Pos.Now.To2D())
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

func (bc *BodyController) PushBack(segment *core.SectorSegment) bool {
	d := segment.DistanceToPoint2(bc.pos2d)
	if d > bc.Body.Size.Now[0]*bc.Body.Size.Now[0]*0.25 {
		return false
	}
	side := segment.WhichSide(bc.pos2d)
	closest := segment.ClosestToPoint(bc.pos2d)
	delta := bc.pos2d.Sub(closest)
	d = delta.Length()
	if d > constants.IntersectEpsilon {
		delta[0] /= d
		delta[1] /= d
	} else {
		delta.From(&segment.Normal)
		delta[0] = -delta[0]
		delta[1] = -delta[1]
		d = 0
	}
	if side > 0 {
		delta.MulSelf(bc.Body.Size.Now[0]*0.5 - d)
	} else {
		log.Printf("PushBack: body is on the front-facing side of segment (%v units)\n", d)
		delta.MulSelf(-bc.Body.Size.Now[0]*0.5 - d)
	}
	// Apply the impulse at the same time
	bc.pos.To2D().AddSelf(delta)
	if d > 0 {
		bc.Body.Vel.Now.To2D().AddSelf(delta)
	}

	return true
}

func (bc *BodyController) findBodySector() {
	var closestSector *core.Sector

	col := ecs.Column[core.Sector](bc.Body.ECS, core.SectorComponentIndex)
	for i := range col.Length {
		sector := col.Value(i)
		if sector.IsPointInside2D(bc.pos2d) {
			closestSector = sector
			break
		}
	}

	if closestSector == nil {
		p := bc.Body.Pos.Now.To2D()
		var closestSeg *core.SectorSegment
		closestDistance2 := math.MaxFloat64
		for i := range col.Length {
			sector := col.Value(i)
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
		bc.Body.Pos.Now[0] = p[0]
		bc.Body.Pos.Now[1] = p[1]
	}

	floorZ, ceilZ := closestSector.ZAt(ecs.DynamicNow, bc.pos2d)
	//log.Printf("F: %v, C:%v\n", floorZ, ceilZ)
	if bc.pos[2]-bc.halfHeight < floorZ || bc.pos[2]+bc.halfHeight > ceilZ {
		//log.Printf("Moved body %v to closest sector and adjusted Z from %v to %v", bc.Body.Entity, p[2], floorZ)
		bc.pos[2] = floorZ + bc.halfHeight
	}
	bc.Enter(closestSector.Entity)
	// Don't mark as collided because this is probably an initialization.
}

func (bc *BodyController) checkBodySegmentCollisions() {
	// See if we need to push back into the current sector.
	for _, segment := range bc.Sector.Segments {
		if segment.AdjacentSector != 0 && segment.PortalIsPassable {
			adj := core.GetSector(bc.Sector.ECS, segment.AdjacentSector)
			// We can still collide with a portal if the heights don't match.
			// If we're within limits, ignore the portal.
			floorZ, ceilZ := adj.ZAt(ecs.DynamicNow, bc.pos2d)
			if bc.pos[2]-bc.halfHeight+bc.Body.MountHeight >= floorZ &&
				bc.pos[2]+bc.halfHeight < ceilZ {
				continue
			}
		}
		if bc.PushBack(segment) {
			bc.collidedSegments = append(bc.collidedSegments, segment)
		}
	}
}

func (bc *BodyController) bodyTeleport() bool {
	for _, segment := range bc.Sector.Segments {
		if !segment.PortalTeleports || segment.AdjacentSegment == nil {
			continue
		}
		d := segment.DistanceToPoint2(bc.pos2d)
		if d > bc.Body.Size.Now[0]*bc.Body.Size.Now[0]*0.25 {
			continue
		}
		side := segment.WhichSide(bc.pos2d)
		if side < 0 {
			// Teleport position
			v := segment.PortalMatrix.Unproject(bc.pos2d)
			v = segment.AdjacentSegment.MirrorPortalMatrix.Project(v)
			bc.Body.Pos.Now[0] = v[0]
			bc.Body.Pos.Now[1] = v[1]
			// Teleport velocity
			trans := *bc.Body.Vel.Now.To2D()
			trans[0] += segment.A[0]
			trans[1] += segment.A[1]
			v = segment.PortalMatrix.Unproject(&trans)
			v = segment.AdjacentSegment.MirrorPortalMatrix.Project(v)
			bc.Body.Vel.Now[0] = v[0] - segment.AdjacentSegment.B[0]
			bc.Body.Vel.Now[1] = v[1] - segment.AdjacentSegment.B[1]
			// Calculate new facing angle
			bc.Body.Angle.Now = bc.Body.Angle.Now -
				math.Atan2(segment.Normal[1], segment.Normal[0])*concepts.Rad2deg +
				math.Atan2(segment.AdjacentSegment.Normal[1], segment.AdjacentSegment.Normal[0])*concepts.Rad2deg + 180
			for bc.Body.Angle.Now > 360 {
				bc.Body.Angle.Now -= 360
			}
			bc.Body.LastEnteredPortal = segment
			bc.Enter(segment.AdjacentSector)
			return true
		}
	}
	return false
}

func (bc *BodyController) bodyExitsSector() {
	// Exit the current sector.
	bc.Exit()

	if bc.bodyTeleport() {
		return
	}

	for _, segment := range bc.Sector.Segments {
		if segment.AdjacentSector == 0 {
			continue
		}
		adj := core.GetSector(bc.Sector.ECS, segment.AdjacentSector)
		floorZ, ceilZ := adj.ZAt(ecs.DynamicNow, bc.pos2d)
		if bc.pos[2]-bc.halfHeight+bc.Body.MountHeight >= floorZ &&
			bc.pos[2]+bc.halfHeight < ceilZ &&
			adj.IsPointInside2D(bc.pos2d) {
			// Hooray, we've handled case 5! Make sure Z is good.
			//log.Printf("Case 5! body = %v in sector %v, floor z = %v\n", p.StringHuman(), adj.Entity, floorZ)
			/*if p[2] < floorZ {
				e.Pos[2] = floorZ
				log.Println("Entity entering adjacent sector is lower than floorZ")
			}*/
			bc.Body.LastEnteredPortal = segment
			bc.Enter(segment.AdjacentSector)
			break
		}
	}

	if bc.Sector == nil {
		// Case 6! This is the worst.
		col := ecs.Column[core.Sector](bc.Body.ECS, core.SectorComponentIndex)
		for i := range col.Length {
			sector := col.Value(i)
			floorZ, ceilZ := sector.ZAt(ecs.DynamicNow, bc.pos2d)
			if bc.pos[2]-bc.halfHeight+bc.Body.MountHeight >= floorZ &&
				bc.pos[2]+bc.halfHeight < ceilZ {
				for _, segment := range sector.Segments {
					if bc.PushBack(segment) {
						bc.collidedSegments = append(bc.collidedSegments, segment)
					}
				}
			}
		}
	}
}

func (bc *BodyController) removeBody(body *core.Body) {
	// TODO: poorly implemented
	sector := body.Sector()
	if sector != nil {
		delete(sector.Bodies, body.Entity)
		body.SectorEntity = 0
		if body == bc.Body {
			bc.Sector = nil
		}
	}
	body.ECS.DetachAll(body.Entity)
}

func (bc *BodyController) resolveCollision(body *core.Body) {
	// Use the right collision response settings
	aResponse := bc.Body.CrBody
	bResponse := body.CrBody
	if behaviors.GetPlayer(body.ECS, body.Entity) != nil {
		aResponse = bc.Body.CrPlayer
	}
	if bc.Player != nil {
		bResponse = body.CrPlayer
	}

	// We scale the push-back by the mass of the body.
	aMass := 0.0
	bMass := 0.0
	if aResponse&core.CollideBounce != 0 || aResponse&core.CollideSeparate != 0 {
		aMass = bc.Body.Mass
	}
	if bResponse&core.CollideBounce != 0 || bResponse&core.CollideSeparate != 0 {
		bMass = body.Mass
	}

	// The code below to bounce rigid bodies is adapted from
	// https://www.myphysicslab.com/engine2D/collision-en.html

	aRadius := bc.Body.Size.Now[0] * 0.5
	bRadius := body.Size.Now[0] * 0.5
	UnitAtoB := bc.pos.Sub(&body.Pos.Now)
	distance := UnitAtoB.Length()
	if distance > constants.IntersectEpsilon {
		UnitAtoB.MulSelf(1.0 / distance)
	} else {
		// The bodies are right on top of each other.
		// Resolve the collision with a random direction vector.
		UnitAtoB[0], UnitAtoB[1] = math.Sincos(rand.Float64() * math.Pi * 2)
	}

	// This code separates inter-penetrating bodies
	distance = aRadius + bRadius - distance
	if aMass > 0 || bMass > 0 {
		aWeight := aMass / (aMass + bMass)
		bWeight := bMass / (aMass + bMass)
		bc.pos[0] += UnitAtoB[0] * distance * aWeight
		bc.pos[1] += UnitAtoB[1] * distance * aWeight
		body.Pos.Now[0] -= UnitAtoB[0] * distance * bWeight
		body.Pos.Now[1] -= UnitAtoB[1] * distance * bWeight
	}

	// Next handle the other cases and check if we need to do a bounce
	if aResponse&core.CollideDeactivate != 0 {
		bc.Body.Active = false
	}
	if aResponse&core.CollideStop != 0 {
		bc.Body.Vel.Now[0] = 0
		bc.Body.Vel.Now[1] = 0
		bc.Body.Vel.Now[2] = 0
	}
	if aResponse&core.CollideRemove != 0 {
		bc.removeBody(bc.Body)
	}
	if bResponse&core.CollideDeactivate != 0 {
		body.Active = false
	}
	if bResponse&core.CollideStop != 0 {
		body.Vel.Now[0] = 0
		body.Vel.Now[1] = 0
		body.Vel.Now[2] = 0
	}
	if bResponse&core.CollideRemove != 0 {
		bc.removeBody(body)
	}

	if aResponse&core.CollideBounce == 0 && bResponse&core.CollideBounce == 0 {
		return
	}

	// These two handle the case when one body is set to only separate, and
	// the other needs to "bounce"
	if aResponse&core.CollideSeparate != 0 {
		aMass = 0.0
	}
	if bResponse&core.CollideSeparate != 0 {
		bMass = 0.0
	}

	r_ap := UnitAtoB.Mul(-aRadius)
	r_bp := UnitAtoB.Mul(bRadius)
	// For now, assume no angular velocity. In the future, this may
	// change.
	//vang_a1, vang_b1 := new(concepts.Vector3), new(concepts.Vector3)
	v_a1, v_b1 := &bc.Body.Vel.Now, &body.Vel.Now
	v_ap1 := v_a1 //.Add(vang_a1.Cross(r_ap))
	v_bp1 := v_b1 //.Add(vang_b1.Cross(r_bp))
	v_p1 := v_ap1.Sub(v_bp1)

	aC := r_ap.Cross(UnitAtoB)
	bC := r_bp.Cross(UnitAtoB)
	// Solid sphere moment of inertia = (2/5)MR^2
	momentA := aMass * aRadius * aRadius * 2.0 / 5.0
	momentB := bMass * bRadius * bRadius * 2.0 / 5.0
	aElasticity := 0.0
	bElasticity := 0.0
	if bc.Body.Elasticity > 0 || body.Elasticity > 0 {
		aElasticity = bc.Body.Elasticity * bc.Body.Elasticity / (bc.Body.Elasticity + body.Elasticity)
		bElasticity = body.Elasticity * body.Elasticity / (bc.Body.Elasticity + body.Elasticity)
	}
	if momentA > 0 && momentB > 0 {
		jA := v_p1.Dot(UnitAtoB) / (1.0/bc.Body.Mass + 1.0/bMass + aC.Dot(aC)/momentA + bC.Dot(bC)/momentB)
		jB := -(1.0 + bElasticity) * jA
		jA = -(1.0 + aElasticity) * jA
		bc.Body.Vel.Now.AddSelf(UnitAtoB.Mul(jA / bMass))
		body.Vel.Now.AddSelf(UnitAtoB.Mul(-jB / bMass))
	} else if momentA > 0 {
		j := -(1.0 + aElasticity) * v_p1.Dot(UnitAtoB) / (1.0/bc.Body.Mass + aC.Dot(aC)/momentA)
		bc.Body.Vel.Now.AddSelf(UnitAtoB.Mul(j / bc.Body.Mass))
	} else if momentB > 0 {
		j := -(1.0 + bElasticity) * v_p1.Dot(UnitAtoB) / (1.0/bMass + bC.Dot(bC)/momentB)
		body.Vel.Now.AddSelf(UnitAtoB.Mul(-j / bMass))
	}
	//fmt.Printf("%v <-> %v = %v\n", bc.Body.String(), body.String(), diff)
}

func (bc *BodyController) getInventoryItem(item *behaviors.InventoryItem) {
	for _, slot := range bc.Player.Inventory {
		if !slot.ValidClasses.Contains(item.Class) {
			continue
		}
		if slot.Count >= slot.Limit {
			bc.Player.Notices.Push("Can't pick up more " + item.Class)
			return
		}
		slot.Count++
		bc.Player.Notices.Push("Picked up a " + item.Class)
	}
}

func (bc *BodyController) bodyBodyCollide(sector *core.Sector) {
	for _, body := range sector.Bodies {
		if body == nil || body == bc.Body || !body.IsActive() {
			continue
		}
		if p := behaviors.GetPlayer(body.ECS, body.Entity); p != nil && p.Spawn {
			// Ignore spawn points
			continue
		}
		// From https://www.myphysicslab.com/engine2D/collision-en.html
		d2 := bc.pos.Dist2(&body.Pos.Now)
		r_a := bc.Body.Size.Now[0] * 0.5
		r_b := body.Size.Now[0] * 0.5
		if d2 < (r_a+r_b)*(r_a+r_b) {
			item := behaviors.GetInventoryItem(body.ECS, body.Entity)
			if item != nil && item.Active && bc.Player != nil {
				bc.getInventoryItem(item)
			}
			bc.resolveCollision(body)
		}
	}
}

func (bc *BodyController) CollideZ() {
	halfHeight := bc.Body.Size.Now[1] * 0.5
	bodyTop := bc.Body.Pos.Now[2] + halfHeight
	floorZ, ceilZ := bc.Sector.ZAt(ecs.DynamicNow, bc.Body.Pos.Now.To2D())

	bc.Body.OnGround = false
	if bc.Sector.Bottom.Target != 0 && bodyTop < floorZ {
		delta := bc.Body.Pos.Now.Sub(&bc.Sector.Center)
		bc.Exit()
		bc.Enter(bc.Sector.Bottom.Target)
		bc.Body.Pos.Now[0] = bc.Sector.Center[0] + delta[0]
		bc.Body.Pos.Now[1] = bc.Sector.Center[1] + delta[1]
		ceilZ = bc.Sector.Top.ZAt(ecs.DynamicNow, bc.Body.Pos.Now.To2D())
		bc.Body.Pos.Now[2] = ceilZ - halfHeight - 1.0
	} else if bc.Sector.Bottom.Target != 0 && bc.Body.Pos.Now[2]-halfHeight <= floorZ && bc.Body.Vel.Now[2] > 0 {
		bc.Body.Vel.Now[2] = constants.PlayerJumpForce
	} else if bc.Sector.Bottom.Target == 0 && bc.Body.Pos.Now[2]-halfHeight <= floorZ {
		dist := bc.Sector.Bottom.Normal[2] * (floorZ - (bc.Body.Pos.Now[2] - halfHeight))
		delta := bc.Sector.Bottom.Normal.Mul(dist)
		// TODO: do this for ceiling too
		c_a := delta.Cross(&bc.Sector.Bottom.Normal)
		// Solid sphere moment of inertia
		moment := bc.Body.Mass * halfHeight * halfHeight * 2.0 / 5.0
		j := -(1.0 + bc.Body.Elasticity) * bc.Body.Vel.Now.Dot(&bc.Sector.Bottom.Normal) / (1.0/bc.Body.Mass + c_a.Dot(c_a)/moment)
		bc.Body.Vel.Now.AddSelf(bc.Sector.Bottom.Normal.Mul(j / bc.Body.Mass))
		bc.Body.Pos.Now.AddSelf(delta)
		bc.Body.OnGround = true
		BodySectorScript(bc.Sector.Bottom.Scripts, bc.Body, bc.Sector)
	}

	if bc.Sector.Top.Target != 0 && bodyTop > ceilZ {
		delta := bc.Body.Pos.Now.Sub(&bc.Sector.Center)
		bc.Exit()
		bc.Enter(bc.Sector.Top.Target)
		bc.Body.Pos.Now[0] = bc.Sector.Center[0] + delta[0]
		bc.Body.Pos.Now[1] = bc.Sector.Center[1] + delta[1]
		floorZ = bc.Sector.Bottom.ZAt(ecs.DynamicNow, bc.Body.Pos.Now.To2D())
		bc.Body.Pos.Now[2] = floorZ + halfHeight + 1.0
	} else if bc.Sector.Top.Target == 0 && bodyTop >= ceilZ {
		dist := -bc.Sector.Top.Normal[2] * (bodyTop - ceilZ + 1.0)
		delta := bc.Sector.Top.Normal.Mul(dist)
		bc.Body.Vel.Now.AddSelf(delta)
		bc.Body.Pos.Now.AddSelf(delta)
		BodySectorScript(bc.Sector.Top.Scripts, bc.Body, bc.Sector)
	}
}

func (bc *BodyController) Collide() {
	// We've got several possibilities we need to handle:
	// 1.   The body has an un-initialized sector, but it's within a sector and doesn't need to be moved.
	// 2.   The body is outside of all sectors. Put it into the nearest sector.
	// 3.   The body is still in its current sector, but it's gotten too close to a wall and needs to be pushed back.
	// 4.   The body is outside of the current sector because it's gone past a wall and needs to be pushed back.
	// 5.   The body is outside of the current sector because it's gone through a portal and needs to change sectors.
	// 6.   The body is outside of the current sector because it's gone through a portal, but it winds up outside of
	//      any sectors and needs to be pushed back into a valid sector using any walls within bounds.
	// 7.   The body has collided with another body nearby. Both bodies need to
	//      be pushed apart.
	// 8.   No collision occured.

	// The central method here is to push back, but the wall that's doing the pushing requires some logic to get.

	// Do 10 collision iterations to avoid spending too much time here.
	for i := 0; i < 10; i++ {
		// Avoid GC thrash
		bc.collidedSegments = bc.collidedSegments[:0]
		// Cases 1 & 2.
		if bc.Sector == nil {
			bc.findBodySector()
		}

		if bc.Body.CrWall != core.CollideNone {
			// Case 3 & 4
			bc.checkBodySegmentCollisions()
		}

		if bc.Sector != nil {
			if !bc.Sector.IsPointInside2D(bc.pos2d) {
				// Cases 5 & 6
				bc.bodyExitsSector()
			}

			bc.CollideZ()
			//		bc.bodyBodyCollide(bc.Sector)
			for _, sector := range bc.Sector.PVS {
				bc.bodyBodyCollide(sector)
			}
		}

		if len(bc.collidedSegments) == 0 {
			if bc.Sector != nil {
				return
			} else {
				continue
			}
		}

		for _, seg := range bc.collidedSegments {
			BodySectorScript(seg.ContactScripts, bc.Body, bc.Sector)
		}

		if bc.Body.CrWall&core.CollideStop != 0 {
			bc.Body.Vel.Now[0] = 0
			bc.Body.Vel.Now[1] = 0
		}
		if bc.Body.CrWall&core.CollideBounce != 0 {
			for _, segment := range bc.collidedSegments {
				n := segment.Normal.To3D(new(concepts.Vector3))
				bc.Body.Vel.Now.SubSelf(n.Mul(2 * bc.Body.Vel.Now.Dot(n)))
			}
		}
		if bc.Body.CrWall&core.CollideRemove != 0 {
			bc.removeBody(bc.Body)
		}
	}
}
