// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package archetypes

import (
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
)

func CreateBasic(db *concepts.EntityComponentDB, componentIndex int) *concepts.EntityRef {
	er := db.RefForNewEntity()
	db.NewAttachedComponent(er.Entity, componentIndex)

	return er
}

func CreateSector(db *concepts.EntityComponentDB) *concepts.EntityRef {
	er := db.RefForNewEntity()
	db.NewAttachedComponent(er.Entity, core.SectorComponentIndex)

	return er
}

func IsLightBody(er concepts.EntityRef) bool {
	return er.Component(core.BodyComponentIndex) != nil &&
		er.Component(core.LightComponentIndex) != nil
}

func CreateLightBody(db *concepts.EntityComponentDB) *concepts.EntityRef {
	er := db.RefForNewEntity()
	body := db.NewAttachedComponent(er.Entity, core.BodyComponentIndex).(*core.Body)
	db.NewAttachedComponent(er.Entity, core.LightComponentIndex)
	body.Size.Original[0] = 2
	body.Size.Original[1] = 2
	body.Size.Reset()

	return er
}

func IsPlayerBody(er concepts.EntityRef) bool {
	return er.Component(core.BodyComponentIndex) != nil &&
		er.Component(behaviors.PlayerComponentIndex) != nil &&
		er.Component(behaviors.AliveComponentIndex) != nil
}

func CreatePlayerBody(db *concepts.EntityComponentDB) *concepts.EntityRef {
	er := db.RefForNewEntity()
	body := db.NewAttachedComponent(er.Entity, core.BodyComponentIndex).(*core.Body)
	_ = db.NewAttachedComponent(er.Entity, behaviors.PlayerComponentIndex).(*behaviors.Player)
	_ = db.NewAttachedComponent(er.Entity, behaviors.AliveComponentIndex).(*behaviors.Alive)

	body.Size.Set(concepts.Vector2{constants.PlayerBoundingRadius * 2, constants.PlayerHeight})
	body.Mass = constants.PlayerMass // kg

	return er
}
