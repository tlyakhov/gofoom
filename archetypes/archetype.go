// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package archetypes

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/ecs"
)

func CreateBasic(db *ecs.ECS, componentID ecs.ComponentID) ecs.Entity {
	e := db.NewEntity()
	db.NewAttachedComponent(e, componentID)
	return e
}

func CreateSector(db *ecs.ECS) ecs.Entity {
	entity := db.NewEntity()
	db.NewAttachedComponent(entity, core.SectorCID)

	return entity
}

func IsLightBody(db *ecs.ECS, e ecs.Entity) bool {
	return db.Component(e, core.BodyCID) != nil &&
		db.Component(e, core.LightCID) != nil
}

func CreateLightBody(db *ecs.ECS) ecs.Entity {
	e := db.NewEntity()
	body := db.NewAttachedComponent(e, core.BodyCID).(*core.Body)
	db.NewAttachedComponent(e, core.LightCID)
	body.Size.Original[0] = 2
	body.Size.Original[1] = 2
	body.Size.ResetToOriginal()

	return e
}

/*func IsPlayerBody(db *ecs.ECS, e ecs.Entity) bool {
	return db.Component(e, core.BodyCID) != nil &&
		db.Component(e, behaviors.PlayerCID) != nil &&
		db.Component(e, behaviors.AliveCID) != nil
}

func CreatePlayerBody(db *ecs.ECS) ecs.Entity {
	e := db.NewEntity()
	body := db.NewAttachedComponent(e, core.BodyCID).(*core.Body)
	body.System = true
	body.Size.SetAll(concepts.Vector2{constants.PlayerBoundingRadius * 2, constants.PlayerHeight})
	body.Mass = constants.PlayerMass // kg
	player := db.NewAttachedComponent(e, behaviors.PlayerCID).(*behaviors.Player)
	player.System = true
	alive := db.NewAttachedComponent(e, behaviors.AliveCID).(*behaviors.Alive)
	alive.System = true

	return e
}*/

func CreateFont(db *ecs.ECS, filename string, name string) ecs.Entity {
	e := db.NewEntity()
	named := db.NewAttachedComponent(e, ecs.NamedCID).(*ecs.Named)
	named.System = true
	named.Name = name
	img := db.NewAttachedComponent(e, materials.ImageCID).(*materials.Image)
	img.System = true
	img.Source = filename
	img.GenerateMipMaps = false
	img.Filter = false
	img.Load()
	sprite := db.NewAttachedComponent(e, materials.SpriteCID).(*materials.Sprite)
	sprite.System = true
	sprite.Rows = 16
	sprite.Cols = 16
	sprite.Image = e
	sprite.Angles = 0
	return e
}
