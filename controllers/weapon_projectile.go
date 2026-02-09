// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"math"
	"math/rand/v2"
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/character"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/inventory"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/ecs"
)

func (wc *WeaponController) fireWeaponProjectile(wcp *inventory.WeaponClassProjectile) {
	if wcp.Projectile == 0 {
		return
	}
	spawner := behaviors.GetSpawner(wcp.Projectile)
	if spawner == nil {
		return
	}

	// TODO: Optimize this
	e := Spawn(spawner)

	if e == 0 {
		return
	}

	flags := ecs.ComponentActive | ecs.ComponentHideEntityInEditor | ecs.ComponentLockedInEditor

	body := core.GetBody(e)
	if body == nil {
		// Each particle has its own position
		body = ecs.NewAttachedComponent(e, core.BodyCID).(*core.Body)
		body.Flags |= flags
	}
	mobile := core.GetMobile(e)
	if mobile == nil {
		// Each particle has its own dynamics
		mobile = ecs.NewAttachedComponent(e, core.MobileCID).(*core.Mobile)
		mobile.Flags |= flags
		mobile.Mass = 0.25
		mobile.CrBody = core.CollideRemove
		mobile.CrPlayer = core.CollideRemove
		mobile.CrWall = core.CollideRemove
	}

	hAngle := wc.Body.Angle.Now + (rand.Float64()-0.5)*wc.Class.Spread
	pitchSpread := (rand.Float64() - 0.5) * wc.Class.Spread
	// TODO: All bodies should probably be able to pitch
	vAngle := pitchSpread
	if p := character.GetPlayer(wc.Body.Entity); p != nil {
		vAngle += p.Pitch
	}

	combinedRadius := wc.Body.Size.Now[0]*0.5 + body.Size.Now[0]*0.5 + 2

	mobile.Vel.Spawn[0] = math.Cos(hAngle*concepts.Deg2rad) * math.Cos(vAngle*concepts.Deg2rad)
	mobile.Vel.Spawn[1] = math.Sin(hAngle*concepts.Deg2rad) * math.Cos(vAngle*concepts.Deg2rad)
	mobile.Vel.Spawn[2] = math.Sin(vAngle * concepts.Deg2rad)
	body.Pos.Spawn[0] = wc.Body.Pos.Now[0] + mobile.Vel.Spawn[0]*combinedRadius
	body.Pos.Spawn[1] = wc.Body.Pos.Now[1] + mobile.Vel.Spawn[1]*combinedRadius
	body.Pos.Spawn[2] = wc.Body.Pos.Now[2] + mobile.Vel.Spawn[2]*combinedRadius
	body.Pos.ResetToSpawn()
	mobile.Vel.Spawn.MulSelf(wcp.Speed)
	mobile.Vel.ResetToSpawn()

	if wc.bodyController.Target(body, e) {
		wc.bodyController.Enter(wc.Body.Sector())
		wc.bodyController.Precompute()
	}
}
