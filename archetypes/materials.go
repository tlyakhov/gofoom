// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package archetypes

import (
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/ecs"
)

func EntityMapIsMaterial(components []ecs.Attachable) bool {
	return len(components) > 0 && (components[materials.ShaderComponentIndex] != nil ||
		components[materials.SpriteComponentIndex] != nil ||
		components[materials.LitComponentIndex] != nil ||
		components[materials.ImageComponentIndex] != nil ||
		components[materials.TextComponentIndex] != nil ||
		components[materials.SolidComponentIndex] != nil)
}

func EntityIsMaterial(db *ecs.ECS, e ecs.Entity) bool {
	return EntityMapIsMaterial(db.AllComponents(e))
}

func AttachableIsMaterial(a ecs.Attachable) bool {
	return EntityIsMaterial(a.GetECS(), a.GetEntity())
}

func CreateBasicMaterial(db *ecs.ECS, textured bool) ecs.Entity {
	e := db.NewEntity()
	named := db.NewAttachedComponent(e, ecs.NamedComponentIndex).(*ecs.Named)
	named.Name = "Material " + e.Format(db)
	db.NewAttachedComponent(e, materials.LitComponentIndex)
	if textured {
		db.NewAttachedComponent(e, materials.ImageComponentIndex)
	} else {
		db.NewAttachedComponent(e, materials.SolidComponentIndex)
	}
	return e
}
