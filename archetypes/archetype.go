// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package archetypes

import (
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
)

func CreateBasic(db *concepts.EntityComponentDB, componentIndex int) concepts.Entity {
	e := db.NewEntity()
	db.NewAttachedComponent(e, componentIndex)
	return e
}

func CreateSector(db *concepts.EntityComponentDB) concepts.Entity {
	entity := db.NewEntity()
	db.NewAttachedComponent(entity, core.SectorComponentIndex)

	return entity
}

func IsLightBody(db *concepts.EntityComponentDB, e concepts.Entity) bool {
	return db.Component(e, core.BodyComponentIndex) != nil &&
		db.Component(e, core.LightComponentIndex) != nil
}

func CreateLightBody(db *concepts.EntityComponentDB) concepts.Entity {
	e := db.NewEntity()
	body := db.NewAttachedComponent(e, core.BodyComponentIndex).(*core.Body)
	db.NewAttachedComponent(e, core.LightComponentIndex)
	body.Size.Original[0] = 2
	body.Size.Original[1] = 2
	body.Size.Reset()

	return e
}

func IsPlayerBody(db *concepts.EntityComponentDB, e concepts.Entity) bool {
	return db.Component(e, core.BodyComponentIndex) != nil &&
		db.Component(e, behaviors.PlayerComponentIndex) != nil &&
		db.Component(e, behaviors.AliveComponentIndex) != nil
}

func CreatePlayerBody(db *concepts.EntityComponentDB) concepts.Entity {
	e := db.NewEntity()
	body := db.NewAttachedComponent(e, core.BodyComponentIndex).(*core.Body)
	_ = db.NewAttachedComponent(e, behaviors.PlayerComponentIndex).(*behaviors.Player)
	_ = db.NewAttachedComponent(e, behaviors.AliveComponentIndex).(*behaviors.Alive)

	body.Size.Set(concepts.Vector2{constants.PlayerBoundingRadius * 2, constants.PlayerHeight})
	body.Mass = constants.PlayerMass // kg

	return e
}
