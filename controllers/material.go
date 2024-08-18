// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/ecs"
)

func DefaultMaterial(db *ecs.ECS) ecs.Entity {
	entity := db.GetEntityByName("Default Material")
	if entity != 0 {
		return entity
	}

	// Otherwise try a random one?
	return db.First(materials.LitComponentIndex).GetEntity()
}
