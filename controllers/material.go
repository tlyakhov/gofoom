// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/ecs"
)

func DefaultMaterial() ecs.Entity {
	entity := ecs.GetEntityByName("Default Material")
	if entity != 0 {
		return entity
	}

	// Otherwise try a random one?
	return ecs.First(materials.LitCID).Base().Entity
}
