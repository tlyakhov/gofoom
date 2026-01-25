// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"math/rand/v2"
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/character"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/inventory"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/components/selection"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
)

func (wc *WeaponController) fireWeaponInstant(instant *inventory.WeaponClassInstant) {
	angle := wc.Body.Angle.Now + (rand.Float64()-0.5)*wc.Class.Spread
	pitchSpread := (rand.Float64() - 0.5) * wc.Class.Spread
	// TODO: All bodies should probably be able to pitch
	pitch := pitchSpread
	if p := character.GetPlayer(wc.Body.Entity); p != nil {
		pitch += p.Pitch
	}
	ray := &concepts.Ray{Start: wc.Body.Pos.Now}
	ray.FromAngleAndLimit(angle, pitch, constants.MaxViewDistance)
	var s *selection.Selectable
	s, wc.hit = Cast(ray, wc.Body.Sector(), wc.Body.Entity, false)
	wc.delta = ray.Delta

	if s == nil {
		return
	}

	// TODO: Account for bullet velocity travel time. Do this by calculating
	// time it would take to hit the thing and delaying the outcome? could be
	// buggy though if the object in question moves
	//log.Printf("Weapon hit! %v[%v] at %v", s.Type, s.Entity, wc.hit.StringHuman(2))
	switch s.Type {
	case selection.SelectableBody:
		if mobile := core.GetMobile(s.Body.Entity); mobile != nil {
			// Push bodies away
			// TODO: Parameterize in Weapon
			mobile.Vel.Now.AddSelf(wc.delta.Mul(3))
		}
		// Hurt anything alive
		if alive := behaviors.GetAlive(s.Body.Entity); alive != nil {
			// TODO: Parameterize CoolDown in WeaponClassInstant?

			alive.Hurt("Weapon "+s.Entity.String(), instant.Damage, 20)
		}
	case selection.SelectableSectorSegment, selection.SelectableHi,
		selection.SelectableLow, selection.SelectableMid,
		selection.SelectableInternalSegment:
		// Make a mark on walls

		// TODO: Include floors and ceilings
		es := &materials.ShaderStage{
			Material:               wc.Class.MarkMaterial,
			IgnoreSurfaceTransform: false,
		}
		// TODO: Fix this
		//es.CFlags = ecs.ComponentInternal
		es.Construct(nil)
		es.Flags = 0
		surf := wc.MarkSurfaceAndTransform(s, &wc.transform)
		surf.ExtraStages = append(surf.ExtraStages, es)
		es.Transform.From(&surf.Transform.Now)
		es.Transform.AffineInverseSelf().MulSelf(&wc.transform)
		wc.updateMarks(inventory.WeaponMark{
			ShaderStage: es,
			Surface:     surf,
		})
	}
}
