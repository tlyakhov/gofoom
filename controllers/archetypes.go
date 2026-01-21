// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"tlyakhov/gofoom/components/audio"
	"tlyakhov/gofoom/components/behaviors"
	_ "tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/inventory"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/ecs"
)

func CreateLightBody() ecs.Entity {
	e := ecs.NewEntity()
	body := ecs.NewAttachedComponent(e, core.BodyCID).(*core.Body)
	ecs.NewAttachedComponent(e, core.LightCID)
	body.Size.Spawn[0] = 2
	body.Size.Spawn[1] = 2
	body.Size.ResetToSpawn()

	return e
}

func CreateFont(filename string, name string) ecs.Entity {
	e := ecs.NewEntity()
	named := ecs.NewAttachedComponent(e, ecs.NamedCID).(*ecs.Named)
	named.Flags |= ecs.EntityInternal
	named.Name = name
	img := ecs.NewAttachedComponent(e, materials.ImageCID).(*materials.Image)
	img.Flags |= ecs.EntityInternal
	img.Source = filename
	img.GenerateMipMaps = false
	img.Filter = false
	img.MarkDirty()
	sprite := ecs.NewAttachedComponent(e, materials.SpriteSheetCID).(*materials.SpriteSheet)
	sprite.Flags |= ecs.EntityInternal
	sprite.Rows = 16
	sprite.Cols = 16
	sprite.Material = e
	sprite.Angles = 0
	ecs.ActAllControllersOneEntity(e, ecs.ControllerPrecompute)
	return e
}

func EntitiesByClass(c string) ecs.EntityTable {
	entitySet := make(ecs.EntityTable, 0)
	cids := make([]ecs.ComponentID, 0)

	switch c {
	case "Sector":
		cids = append(cids, core.SectorCID)
	case "Material":
		cids = append(cids, materials.ShaderCID, materials.SpriteSheetCID,
			materials.ImageCID, materials.TextCID, materials.SolidCID)
	case "Action":
		cids = append(cids, behaviors.ActionFaceCID, behaviors.ActionWaypointCID,
			behaviors.ActionJumpCID, behaviors.ActionFireCID, behaviors.ActionTransitionCID)
	case "Weapon":
		cids = append(cids, inventory.WeaponClassCID)
	case "Sound":
		cids = append(cids, audio.SoundCID)
	case "Spawner":
		cids = append(cids, behaviors.SpawnerCID)
	}
	for _, cid := range cids {
		arena := ecs.ArenaByID(cid)
		for i := range arena.Cap() {
			if a := arena.Component(i); a != nil {
				e := a.Base().Entity
				if entitySet.Contains(e) {
					continue
				}
				entitySet.Set(e)
			}
		}
	}
	return entitySet
}

func DefaultMaterial() ecs.Entity {
	set := EntitiesByClass("Material")
	var first ecs.Entity
	for _, e := range set {
		if e == 0 {
			continue
		}
		if first == 0 {
			first = e
		}
		if named := ecs.GetNamed(e); named != nil && named.Name == "Default" {
			return e
		}
	}

	// Otherwise try a random one?
	return first
}
