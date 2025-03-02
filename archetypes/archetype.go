// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package archetypes

import (
	_ "tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/ecs"
)

func CreateLightBody(db *ecs.ECS) ecs.Entity {
	e := db.NewEntity()
	body := db.NewAttachedComponent(e, core.BodyCID).(*core.Body)
	db.NewAttachedComponent(e, core.LightCID)
	body.Size.Spawn[0] = 2
	body.Size.Spawn[1] = 2
	body.Size.ResetToSpawn()

	return e
}

func CreateFont(db *ecs.ECS, filename string, name string) ecs.Entity {
	e := db.NewEntity()
	named := db.NewAttachedComponent(e, ecs.NamedCID).(*ecs.Named)
	named.Flags = ecs.ComponentInternal
	named.Name = name
	img := db.NewAttachedComponent(e, materials.ImageCID).(*materials.Image)
	img.Flags = ecs.ComponentInternal
	img.Source = filename
	img.GenerateMipMaps = false
	img.Filter = false
	img.Load()
	sprite := db.NewAttachedComponent(e, materials.SpriteSheetCID).(*materials.SpriteSheet)
	sprite.Flags = ecs.ComponentInternal
	sprite.Rows = 16
	sprite.Cols = 16
	sprite.Material = e
	sprite.Angles = 0
	db.ActAllControllersOneEntity(e, ecs.ControllerRecalculate)
	return e
}
