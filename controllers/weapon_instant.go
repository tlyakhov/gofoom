// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"math/rand/v2"
	"tlyakhov/gofoom/components/character"
	"tlyakhov/gofoom/components/selection"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
)

func (wc *WeaponController) fireWeaponInstant() *selection.Selectable {
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
	return s
}
