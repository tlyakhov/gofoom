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
	"tlyakhov/gofoom/dynamic"
	"tlyakhov/gofoom/ecs"
)

func BodySectorScript(scripts []*core.Script, body *core.Body, sector *core.Sector) {
	for _, script := range scripts {
		script.Vars["body"] = body
		script.Vars["sector"] = sector
		script.Act()
	}
}

func (mc *MobileController) PushBack(segment *core.SectorSegment) bool {
	d := segment.DistanceToPoint2(mc.pos2d)
	if d > mc.Body.Size.Now[0]*mc.Body.Size.Now[0]*0.25 {
		return false
	}
	side := segment.WhichSide(mc.pos2d)
	closest := segment.ClosestToPoint(mc.pos2d)
	delta := mc.pos2d.Sub(closest)
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
		delta.MulSelf(mc.Body.Size.Now[0]*0.5 - d)
	} else {
		log.Printf("PushBack: body is on the front-facing side of segment (%v units)\n", d)
		delta.MulSelf(-mc.Body.Size.Now[0]*0.5 - d)
	}
	// Apply the impulse at the same time
	mc.pos.To2D().AddSelf(delta)
	if d > 0 {
		mc.Vel.Now.To2D().AddSelf(delta)
	}

	return true
}

func (mc *MobileController) checkBodySegmentCollisions() {
	// See if we need to push back into the current sector.
	for _, segment := range mc.Sector.Segments {
		if segment.AdjacentSector != 0 && segment.PortalIsPassable {
			adj := core.GetSector(mc.Sector.ECS, segment.AdjacentSector)
			// We can still collide with a portal if the heights don't match.
			// If we're within limits, ignore the portal.
			floorZ, ceilZ := adj.ZAt(dynamic.DynamicNow, mc.pos2d)
			if mc.pos[2]-mc.halfHeight+mc.MountHeight >= floorZ &&
				mc.pos[2]+mc.halfHeight < ceilZ {
				continue
			}
		}
		if mc.PushBack(segment) {
			mc.collidedSegments = append(mc.collidedSegments, segment)
		}
	}
}

func (mc *MobileController) bodyTeleport() bool {
	for _, segment := range mc.Sector.Segments {
		if !segment.PortalTeleports || segment.AdjacentSegment == nil {
			continue
		}
		d := segment.DistanceToPoint2(mc.pos2d)
		if d > mc.Body.Size.Now[0]*mc.Body.Size.Now[0]*0.25 {
			continue
		}
		side := segment.WhichSide(mc.pos2d)
		if side < 0 {
			// Teleport position
			v := segment.PortalMatrix.Unproject(mc.pos2d)
			v = segment.AdjacentSegment.MirrorPortalMatrix.Project(v)
			mc.Body.Pos.Now[0] = v[0]
			mc.Body.Pos.Now[1] = v[1]
			// Teleport velocity
			trans := *mc.Vel.Now.To2D()
			trans[0] += segment.A[0]
			trans[1] += segment.A[1]
			v = segment.PortalMatrix.Unproject(&trans)
			v = segment.AdjacentSegment.MirrorPortalMatrix.Project(v)
			mc.Vel.Now[0] = v[0] - segment.AdjacentSegment.B[0]
			mc.Vel.Now[1] = v[1] - segment.AdjacentSegment.B[1]
			// Calculate new facing angle
			mc.Body.Angle.Now = mc.Body.Angle.Now -
				math.Atan2(segment.Normal[1], segment.Normal[0])*concepts.Rad2deg +
				math.Atan2(segment.AdjacentSegment.Normal[1], segment.AdjacentSegment.Normal[0])*concepts.Rad2deg + 180
			for mc.Body.Angle.Now > 360 {
				mc.Body.Angle.Now -= 360
			}
			mc.Enter(segment.AdjacentSector)
			return true
		}
	}
	return false
}

func (mc *MobileController) bodyExitsSector() {
	// Exit the current sector.
	mc.Exit()

	if mc.bodyTeleport() {
		return
	}

	for _, segment := range mc.Sector.Segments {
		if segment.AdjacentSector == 0 {
			continue
		}
		adj := core.GetSector(mc.Sector.ECS, segment.AdjacentSector)
		floorZ, ceilZ := adj.ZAt(dynamic.DynamicNow, mc.pos2d)
		if mc.pos[2]-mc.halfHeight+mc.MountHeight >= floorZ &&
			mc.pos[2]+mc.halfHeight < ceilZ &&
			adj.IsPointInside2D(mc.pos2d) {
			// Hooray, we've handled case 5! Make sure Z is good.
			//log.Printf("Case 5! body = %v in sector %v, floor z = %v\n", p.StringHuman(), adj.Entity, floorZ)
			/*if p[2] < floorZ {
				e.Pos[2] = floorZ
				log.Println("Entity entering adjacent sector is lower than floorZ")
			}*/
			mc.Enter(segment.AdjacentSector)
			break
		}
	}

	if mc.Sector == nil {
		// Case 6! This is the worst.
		col := ecs.ColumnFor[core.Sector](mc.Body.ECS, core.SectorCID)
		for i := range col.Length {
			sector := col.Value(i)
			floorZ, ceilZ := sector.ZAt(dynamic.DynamicNow, mc.pos2d)
			if mc.pos[2]-mc.halfHeight+mc.MountHeight >= floorZ &&
				mc.pos[2]+mc.halfHeight < ceilZ {
				for _, segment := range sector.Segments {
					if mc.PushBack(segment) {
						mc.collidedSegments = append(mc.collidedSegments, segment)
					}
				}
			}
		}
	}
}

func (mc *MobileController) removeBody(body *core.Body) {
	// TODO: poorly implemented
	sector := body.Sector()
	if sector != nil {
		delete(sector.Bodies, body.Entity)
		body.SectorEntity = 0
		if body == mc.Body {
			mc.Sector = nil
		}
	}
	body.ECS.DetachAll(body.Entity)
}

func (mc *MobileController) resolveCollision(other *core.Mobile, otherBody *core.Body) {
	// Use the right collision response settings
	aResponse := mc.CrBody
	bResponse := core.CollideNone
	otherElasticity := 0.0
	if other != nil {
		bResponse = other.CrBody
		otherElasticity = other.Elasticity
		if behaviors.GetPlayer(other.ECS, other.Entity) != nil {
			aResponse = mc.CrPlayer
		}
		if mc.Player != nil {
			bResponse = other.CrPlayer
		}
	}

	// We scale the push-back by the mass of the body.
	aMass := 0.0
	bMass := 0.0
	if aResponse&core.CollideBounce != 0 || aResponse&core.CollideSeparate != 0 {
		aMass = mc.Mass
	}
	if bResponse&core.CollideBounce != 0 || bResponse&core.CollideSeparate != 0 {
		bMass = other.Mass
	}

	// The code below to bounce rigid bodies is adapted from
	// https://www.myphysicslab.com/engine2D/collision-en.html

	aRadius := mc.Body.Size.Now[0] * 0.5
	bRadius := otherBody.Size.Now[0] * 0.5
	UnitAtoB := mc.pos.Sub(&otherBody.Pos.Now)
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
		mc.pos[0] += UnitAtoB[0] * distance * aWeight
		mc.pos[1] += UnitAtoB[1] * distance * aWeight
		otherBody.Pos.Now[0] -= UnitAtoB[0] * distance * bWeight
		otherBody.Pos.Now[1] -= UnitAtoB[1] * distance * bWeight
	}

	// Next handle the other cases and check if we need to do a bounce
	if aResponse&core.CollideDeactivate != 0 {
		mc.Body.Active = false
	}
	if aResponse&core.CollideStop != 0 {
		mc.Vel.Now[0] = 0
		mc.Vel.Now[1] = 0
		mc.Vel.Now[2] = 0
	}
	if aResponse&core.CollideRemove != 0 {
		mc.removeBody(mc.Body)
	}

	if bResponse&core.CollideDeactivate != 0 {
		other.Active = false
	}
	if bResponse&core.CollideStop != 0 {
		other.Vel.Now[0] = 0
		other.Vel.Now[1] = 0
		other.Vel.Now[2] = 0
	}
	if bResponse&core.CollideRemove != 0 {
		mc.removeBody(otherBody)
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
	v_a1, v_b1 := &mc.Vel.Now, &other.Vel.Now
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
	if mc.Elasticity > 0 || otherElasticity > 0 {
		aElasticity = mc.Elasticity * mc.Elasticity / (mc.Elasticity + otherElasticity)
		bElasticity = otherElasticity * otherElasticity / (mc.Elasticity + otherElasticity)
	}
	if momentA > 0 && momentB > 0 {
		jA := v_p1.Dot(UnitAtoB) / (1.0/mc.Mass + 1.0/bMass + aC.Dot(aC)/momentA + bC.Dot(bC)/momentB)
		jB := -(1.0 + bElasticity) * jA
		jA = -(1.0 + aElasticity) * jA
		mc.Vel.Now.AddSelf(UnitAtoB.Mul(jA / bMass))
		other.Vel.Now.AddSelf(UnitAtoB.Mul(-jB / bMass))
	} else if momentA > 0 {
		j := -(1.0 + aElasticity) * v_p1.Dot(UnitAtoB) / (1.0/mc.Mass + aC.Dot(aC)/momentA)
		mc.Vel.Now.AddSelf(UnitAtoB.Mul(j / mc.Mass))
	} else if momentB > 0 {
		j := -(1.0 + bElasticity) * v_p1.Dot(UnitAtoB) / (1.0/bMass + bC.Dot(bC)/momentB)
		other.Vel.Now.AddSelf(UnitAtoB.Mul(-j / bMass))
	}
	//fmt.Printf("%v <-> %v = %v\n", mc.Body.String(), body.String(), diff)
}

func (mc *MobileController) getInventoryItem(item *behaviors.InventoryItem) {
	for _, slot := range mc.Player.Inventory {
		if !slot.ValidClasses.Contains(item.Class) {
			continue
		}
		if slot.Count >= slot.Limit {
			mc.Player.Notices.Push("Can't pick up more " + item.Class)
			return
		}
		slot.Count++
		mc.Player.Notices.Push("Picked up a " + item.Class)
	}
}

func (mc *MobileController) bodyBodyCollide(sector *core.Sector) {
	for _, body := range sector.Bodies {
		if body == nil || body == mc.Body || !body.IsActive() {
			continue
		}
		if p := behaviors.GetPlayer(body.ECS, body.Entity); p != nil && p.Spawn {
			// Ignore spawn points
			continue
		}
		// From https://www.myphysicslab.com/engine2D/collision-en.html
		d2 := mc.pos.Dist2(&body.Pos.Now)
		r_a := mc.Body.Size.Now[0] * 0.5
		r_b := body.Size.Now[0] * 0.5
		if d2 < (r_a+r_b)*(r_a+r_b) {
			item := behaviors.GetInventoryItem(body.ECS, body.Entity)
			if item != nil && item.Active && mc.Player != nil {
				mc.getInventoryItem(item)
			}
			mc.resolveCollision(core.GetMobile(body.ECS, body.Entity), body)
		}
	}
}

func (mc *MobileController) CollideZ() {
	halfHeight := mc.Body.Size.Now[1] * 0.5
	bodyTop := mc.Body.Pos.Now[2] + halfHeight
	floorZ, ceilZ := mc.Sector.ZAt(dynamic.DynamicNow, mc.Body.Pos.Now.To2D())

	mc.Body.OnGround = false
	if mc.Sector.Bottom.Target != 0 && bodyTop < floorZ {
		delta := mc.Body.Pos.Now.Sub(&mc.Sector.Center)
		mc.Exit()
		mc.Enter(mc.Sector.Bottom.Target)
		mc.Body.Pos.Now[0] = mc.Sector.Center[0] + delta[0]
		mc.Body.Pos.Now[1] = mc.Sector.Center[1] + delta[1]
		ceilZ = mc.Sector.Top.ZAt(dynamic.DynamicNow, mc.Body.Pos.Now.To2D())
		mc.Body.Pos.Now[2] = ceilZ - halfHeight - 1.0
	} else if mc.Sector.Bottom.Target != 0 && mc.Body.Pos.Now[2]-halfHeight <= floorZ && mc.Vel.Now[2] > 0 {
		mc.Vel.Now[2] = constants.PlayerJumpForce
	} else if mc.Sector.Bottom.Target == 0 && mc.Body.Pos.Now[2]-halfHeight <= floorZ {
		dist := mc.Sector.Bottom.Normal[2] * (floorZ - (mc.Body.Pos.Now[2] - halfHeight))
		delta := mc.Sector.Bottom.Normal.Mul(dist)
		// TODO: do this for ceiling too
		c_a := delta.Cross(&mc.Sector.Bottom.Normal)
		// Solid sphere moment of inertia
		moment := mc.Mass * halfHeight * halfHeight * 2.0 / 5.0
		j := -(1.0 + mc.Elasticity) * mc.Vel.Now.Dot(&mc.Sector.Bottom.Normal) / (1.0/mc.Mass + c_a.Dot(c_a)/moment)
		mc.Vel.Now.AddSelf(mc.Sector.Bottom.Normal.Mul(j / mc.Mass))
		mc.Body.Pos.Now.AddSelf(delta)
		mc.Body.OnGround = true
		BodySectorScript(mc.Sector.Bottom.Scripts, mc.Body, mc.Sector)
	}

	if mc.Sector.Top.Target != 0 && bodyTop > ceilZ {
		delta := mc.Body.Pos.Now.Sub(&mc.Sector.Center)
		mc.Exit()
		mc.Enter(mc.Sector.Top.Target)
		mc.Body.Pos.Now[0] = mc.Sector.Center[0] + delta[0]
		mc.Body.Pos.Now[1] = mc.Sector.Center[1] + delta[1]
		floorZ = mc.Sector.Bottom.ZAt(dynamic.DynamicNow, mc.Body.Pos.Now.To2D())
		mc.Body.Pos.Now[2] = floorZ + halfHeight + 1.0
	} else if mc.Sector.Top.Target == 0 && bodyTop >= ceilZ {
		dist := -mc.Sector.Top.Normal[2] * (bodyTop - ceilZ + 1.0)
		delta := mc.Sector.Top.Normal.Mul(dist)
		mc.Vel.Now.AddSelf(delta)
		mc.Body.Pos.Now.AddSelf(delta)
		BodySectorScript(mc.Sector.Top.Scripts, mc.Body, mc.Sector)
	}
}

func (mc *MobileController) Collide() {
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
		mc.collidedSegments = mc.collidedSegments[:0]
		// Cases 1 & 2.
		if mc.Sector == nil {
			mc.findBodySector()
		}

		if mc.CrWall != core.CollideNone {
			// Case 3 & 4
			mc.checkBodySegmentCollisions()
		}

		if mc.Sector != nil {
			if !mc.Sector.IsPointInside2D(mc.pos2d) {
				// Cases 5 & 6
				mc.bodyExitsSector()
			}

			mc.CollideZ()
			//		mc.bodyBodyCollide(mc.Sector)
			for _, sector := range mc.Sector.PVS {
				mc.bodyBodyCollide(sector)
			}
		}

		if len(mc.collidedSegments) == 0 {
			if mc.Sector != nil {
				return
			} else {
				continue
			}
		}

		for _, seg := range mc.collidedSegments {
			BodySectorScript(seg.ContactScripts, mc.Body, mc.Sector)
		}

		if mc.CrWall&core.CollideStop != 0 {
			mc.Vel.Now[0] = 0
			mc.Vel.Now[1] = 0
		}
		if mc.CrWall&core.CollideBounce != 0 {
			for _, segment := range mc.collidedSegments {
				n := segment.Normal.To3D(new(concepts.Vector3))
				mc.Vel.Now.SubSelf(n.Mul(2 * mc.Vel.Now.Dot(n)))
			}
		}
		if mc.CrWall&core.CollideRemove != 0 {
			mc.removeBody(mc.Body)
		}
	}
}
