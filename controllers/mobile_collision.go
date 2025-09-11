// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"log"
	"math"
	"math/rand/v2"
	"tlyakhov/gofoom/components/character"
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

func (mc *MobileController) PushBack(sector *core.Sector, segment *core.Segment, inner bool) bool {
	d := segment.DistanceToPointSq(mc.pos2d)
	if d > mc.Body.Size.Now[0]*mc.Body.Size.Now[0]*0.25 {
		return false
	}

	// What we are trying to do is ensure the body is always at least half of
	// Body.Size.Now[0] distance away from a wall, but when we push away, we do
	// it tangent to the wall so the body can slide along walls. To achieve
	// this, we first create the (unit length) vector `delta`, which points from
	// wall->body. Then, we scale that by -d to move the player to align with
	// the wall, and +mc.Body.Size.Now[0]*0.5 to get them the right distance
	// away.

	// side > 0 if the body is in the direction of the normal, or < 0 if on
	// opposite side.
	side := segment.WhichSide(mc.pos2d)
	// Closest point to body along segment.
	closest := segment.ClosestToPoint(mc.pos2d)
	delta := mc.pos2d.Sub(closest)
	// Store the distance away, then normalize the delta vector.
	d = delta.Length()
	if d > constants.IntersectEpsilon {
		delta[0] /= d
		delta[1] /= d
	} else {
		// If d is too close, use the segment normal, but a d of zero!
		delta.From(&segment.Normal)
		delta[0] = -delta[0]
		delta[1] = -delta[1]
		d = 0
		side = 1
	}

	// `inner` is only true if we are pushing backwards from the OUTSIDE of the
	// inner sector.
	if inner {
		side *= -1
	}

	// For debugging collisions with segments:
	//log.Printf("PushBack: sector=%v,p=%v, closest=%v, side=%v, delta=%v, d=%.2f, xsize=%.2f",
	//	sector.Entity.String(), mc.pos2d.StringHuman(), closest.StringHuman(), side >= 0, delta.StringHuman(), d, mc.Body.Size.Now[0])

	if side > 0 {
		delta.MulSelf(-d + mc.Body.Size.Now[0]*0.5)
	} else {
		delta.MulSelf(-d - mc.Body.Size.Now[0]*0.5)
	}

	if side < 0 {
		log.Printf("PushBack: body is on the front-facing side of segment (%v units)\n", d)
	}
	// Apply the impulse at the same time
	mc.pos.To2D().AddSelf(delta)
	if d > 0 {
		mc.Vel.Now.To2D().AddSelf(delta)
	}
	mc.collidedSegments = append(mc.collidedSegments, segment)
	return true
}

func (mc *MobileController) checkBodySegmentCollisions() {
	var adj *core.Sector
	// See if we need to push back into the current sector.
	for _, segment := range mc.Sector.Segments {
		if segment.AdjacentSegment != nil && segment.PortalIsPassable {
			adj = segment.AdjacentSegment.Sector
		} else if !mc.Sector.Outer.Empty() {
			adj = mc.Sector.OuterAt(mc.pos2d)
		} else {
			adj = nil
		}
		if adj != nil && mc.sectorEnterable(adj) {
			continue
		}
		mc.PushBack(mc.Sector, &segment.Segment, false)
	}
}

func (mc *MobileController) bodyTeleport() bool {
	for _, segment := range mc.Sector.Segments {
		if !segment.PortalTeleports || segment.AdjacentSegment == nil {
			continue
		}
		d := segment.DistanceToPointSq(mc.pos2d)
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
			mc.Body.Pos.Prev[0] = v[0]
			mc.Body.Pos.Prev[1] = v[1]
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
			mc.Body.Angle.Now = concepts.NormalizeAngle(mc.Body.Angle.Now)
			mc.Body.Angle.Prev = mc.Body.Angle.Now
			mc.Enter(core.GetSector(segment.AdjacentSector))
			return true
		}
	}
	return false
}

func (mc *MobileController) sectorEnterable(test *core.Sector) bool {
	if test == nil {
		return false
	}
	floorZ, ceilZ := test.ZAt(dynamic.Now, mc.pos2d)
	if mc.pos[2]-mc.halfHeight+mc.MountHeight >= floorZ &&
		mc.pos[2]+mc.halfHeight < ceilZ {
		return true
	}
	return false
}

func (mc *MobileController) checkInnerSectors(test *core.Sector) *core.Sector {
	var inner *core.Sector
	result := test
	for _, e := range test.Inner {
		if e == 0 {
			continue
		}
		if inner = core.GetSector(e); inner == nil {
			continue
		}
		if mc.sectorEnterable(inner) {
			if inner.IsPointInside2D(mc.pos2d) {
				return mc.checkInnerSectors(inner)
			}
		} else {
			for _, seg := range inner.Segments {
				mc.PushBack(inner, &seg.Segment, true)
			}
		}
	}
	return result
}

func (mc *MobileController) bodyExitsSector() {
	previous := mc.Sector
	// Exit the current sector.
	mc.Exit()

	if mc.bodyTeleport() {
		return
	}

	for _, segment := range mc.Sector.Segments {
		if segment.AdjacentSector == 0 {
			continue
		}

		if adj := core.GetSector(segment.AdjacentSector); mc.sectorEnterable(adj) && adj.IsPointInside2D(mc.pos2d) {
			mc.Enter(adj)
			return
		}

	}

	outer := previous.OuterAt(mc.pos2d)
	if mc.sectorEnterable(outer) {
		mc.Enter(outer)
		return
	}

	// Case 7! This is the worst.
	arena := ecs.ArenaFor[core.Sector](core.SectorCID)
	for i := range arena.Cap() {
		sector := arena.Value(i)
		if sector == nil {
			continue
		}
		floorZ, ceilZ := sector.ZAt(dynamic.Now, mc.pos2d)
		if mc.pos[2]-mc.halfHeight+mc.MountHeight >= floorZ &&
			mc.pos[2]+mc.halfHeight < ceilZ {
			for _, segment := range sector.Segments {
				mc.PushBack(sector, &segment.Segment, false)
			}
		}
	}

}

func (mc *MobileController) resolveCollision(bMobile *core.Mobile, bBody *core.Body) {
	// Use the right collision response settings
	otherVel := &concepts.Vector3{}
	aResponse := mc.CrBody
	bResponse := core.CollideNone
	otherElasticity := 0.0
	if bMobile != nil {
		bResponse = bMobile.CrBody
		otherElasticity = bMobile.Elasticity
		otherVel = &bMobile.Vel.Now
		if character.GetPlayer(bMobile.Entity) != nil {
			aResponse = mc.CrPlayer
		}
		if mc.Player != nil {
			bResponse = bMobile.CrPlayer
		}
	}

	// We scale the push-back by the mass of the body.
	aMass := 0.0
	bMass := 0.0
	if aResponse&core.CollideBounce != 0 || aResponse&core.CollideSeparate != 0 {
		aMass = mc.Mass
	}
	if bResponse&core.CollideBounce != 0 || bResponse&core.CollideSeparate != 0 {
		bMass = bMobile.Mass
	}

	// The code below to bounce rigid bodies is adapted from
	// https://www.myphysicslab.com/engine2D/collision-en.html

	aRadius := mc.Body.Size.Now[0] * 0.5
	bRadius := bBody.Size.Now[0] * 0.5
	UnitAtoB := mc.pos.Sub(&bBody.Pos.Now)
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
		bBody.Pos.Now[0] -= UnitAtoB[0] * distance * bWeight
		bBody.Pos.Now[1] -= UnitAtoB[1] * distance * bWeight
	}

	// Next handle the other cases and check if we need to do a bounce
	if aResponse&core.CollideDeactivate != 0 {
		mc.Body.Flags &= ^ecs.ComponentActive
	}
	if aResponse&core.CollideStop != 0 {
		mc.Vel.Now[0] = 0
		mc.Vel.Now[1] = 0
		mc.Vel.Now[2] = 0
	}
	if aResponse&core.CollideRemove != 0 {
		mc.Sector = nil
		ecs.Delete(mc.Body.Entity)
	}

	if bResponse&core.CollideDeactivate != 0 {
		bMobile.Flags &= ^ecs.ComponentActive
	}
	if bResponse&core.CollideStop != 0 {
		bMobile.Vel.Now[0] = 0
		bMobile.Vel.Now[1] = 0
		bMobile.Vel.Now[2] = 0
	}
	if bResponse&core.CollideRemove != 0 {
		ecs.Delete(bBody.Entity)
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
	v_a1, v_b1 := &mc.Vel.Now, otherVel
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
		bMobile.Vel.Now.AddSelf(UnitAtoB.Mul(-jB / bMass))
	} else if momentA > 0 {
		j := -(1.0 + aElasticity) * v_p1.Dot(UnitAtoB) / (1.0/mc.Mass + aC.Dot(aC)/momentA)
		mc.Vel.Now.AddSelf(UnitAtoB.Mul(j / mc.Mass))
	} else if momentB > 0 {
		j := -(1.0 + bElasticity) * v_p1.Dot(UnitAtoB) / (1.0/bMass + bC.Dot(bC)/momentB)
		bMobile.Vel.Now.AddSelf(UnitAtoB.Mul(-j / bMass))
	}
	//fmt.Printf("%v <-> %v = %v\n", mc.Body.String(), body.String(), diff)
}

func (mc *MobileController) bodyBodyCollide() {
	mc.tree.Root.RangeCircle(mc.Body.Pos.Now.To2D(), mc.Body.Size.Now[0]*0.5, func(body *core.Body) bool {
		if !body.IsActive() || body == mc.Body {
			return true
		}
		mobile := core.GetMobile(body.Entity)
		if mobile == nil || !mobile.IsActive() {
			return true
		}

		if p := character.GetPlayer(body.Entity); p != nil && p.Spawn {
			// Ignore spawn points
			return true
		}
		// From https://www.myphysicslab.com/engine2D/collision-en.html
		d2 := mc.pos.DistSq(&body.Pos.Now)
		r_a := mc.Body.Size.Now[0] * 0.5
		r_b := body.Size.Now[0] * 0.5
		if d2 < (r_a+r_b)*(r_a+r_b) {
			mc.resolveCollision(mobile, body)
		}

		return true
	})
}

func (mc *MobileController) CollideZ() {
	halfHeight := mc.Body.Size.Now[1] * 0.5
	bodyTop := mc.Body.Pos.Now[2] + halfHeight
	floorZ, ceilZ := mc.Sector.ZAt(dynamic.Now, mc.Body.Pos.Now.To2D())

	mc.Body.OnGround = false
	if mc.Sector.Bottom.Target != 0 && bodyTop < floorZ {
		delta := mc.Body.Pos.Now.Sub(&mc.Sector.Center)
		mc.Exit()
		mc.Enter(core.GetSector(mc.Sector.Bottom.Target))
		mc.Body.Pos.Now[0] = mc.Sector.Center[0] + delta[0]
		mc.Body.Pos.Now[1] = mc.Sector.Center[1] + delta[1]
		ceilZ = mc.Sector.Top.ZAt(dynamic.Now, mc.Body.Pos.Now.To2D())
		mc.Body.Pos.Now[2] = ceilZ - halfHeight - 1.0
	} else if mc.Sector.Bottom.Target != 0 && mc.Body.Pos.Now[2]-halfHeight <= floorZ && mc.Vel.Now[2] > 0 {
		mc.Vel.Now[2] = constants.PlayerJumpForce
	} else if mc.Sector.Bottom.Target == 0 && mc.Body.Pos.Now[2]-halfHeight <= floorZ {
		dist := mc.Sector.Bottom.Normal[2] * (floorZ - (mc.Body.Pos.Now[2] - halfHeight))
		delta := mc.Sector.Bottom.Normal.Mul(dist)
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
		mc.Enter(core.GetSector(mc.Sector.Top.Target))
		mc.Body.Pos.Now[0] = mc.Sector.Center[0] + delta[0]
		mc.Body.Pos.Now[1] = mc.Sector.Center[1] + delta[1]
		floorZ = mc.Sector.Bottom.ZAt(dynamic.Now, mc.Body.Pos.Now.To2D())
		mc.Body.Pos.Now[2] = floorZ + halfHeight + 1.0
	} else if mc.Sector.Top.Target == 0 && bodyTop >= ceilZ {
		dist := -mc.Sector.Top.Normal[2] * (bodyTop - ceilZ + 1.0)
		delta := mc.Sector.Top.Normal.Mul(dist)
		c_a := delta.Cross(&mc.Sector.Top.Normal)
		// Solid sphere moment of inertia
		moment := mc.Mass * halfHeight * halfHeight * 2.0 / 5.0
		j := -(1.0 + mc.Elasticity) * mc.Vel.Now.Dot(&mc.Sector.Top.Normal) / (1.0/mc.Mass + c_a.Dot(c_a)/moment)
		mc.Vel.Now.AddSelf(mc.Sector.Top.Normal.Mul(j / mc.Mass))
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
	// 5.   The body is outside of the current sector because it's gone through
	//      a portal and needs to change sectors.
	// 6.   The body is outside of the current sector because it's gone to the
	//      outer sector.
	// 7.   The body is outside of the current sector because it's gone through a portal, but it winds up outside of
	//      any sectors and needs to be pushed back into a valid sector using any walls within bounds.
	// 8.   The body has collided with another body nearby. Both bodies need to
	//      be pushed apart.
	// 9.   No collision occured.

	// The central method here is to push back, but the wall that's doing the pushing requires some logic to get.

	// Do 10 collision iterations to avoid spending too much time here.
	for range 10 {
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
			inner := mc.checkInnerSectors(mc.Sector)
			if inner != mc.Sector {
				mc.Exit()
				mc.Enter(inner)
			}

			if !mc.Sector.IsPointInside2D(mc.pos2d) {
				// Cases 5, 6, and 7
				mc.bodyExitsSector()
			}

			mc.CollideZ()
		}

		mc.bodyBodyCollide()

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
			mc.Sector = nil
			ecs.Delete(mc.Body.Entity)
		}
	}
}
