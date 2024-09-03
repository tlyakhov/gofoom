// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package archetypes

import (
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/ecs"
)

func EntityMapIsMaterial(components []ecs.Attachable) bool {
	return len(components) > 0 && (components[materials.ShaderCID>>16] != nil ||
		components[materials.SpriteCID>>16] != nil ||
		components[materials.LitCID>>16] != nil ||
		components[materials.ImageCID>>16] != nil ||
		components[materials.TextCID>>16] != nil ||
		components[materials.SolidCID>>16] != nil)
}

func EntityIsMaterial(db *ecs.ECS, e ecs.Entity) bool {
	return EntityMapIsMaterial(db.AllComponents(e))
}

func AttachableIsMaterial(a ecs.Attachable) bool {
	return EntityIsMaterial(a.GetECS(), a.GetEntity())
}

func CreateBasicMaterial(db *ecs.ECS, textured bool) ecs.Entity {
	e := db.NewEntity()
	named := db.NewAttachedComponent(e, ecs.NamedCID).(*ecs.Named)
	named.Name = "Material " + e.Format(db)
	db.NewAttachedComponent(e, materials.LitCID)
	if textured {
		db.NewAttachedComponent(e, materials.ImageCID)
	} else {
		db.NewAttachedComponent(e, materials.SolidCID)
	}
	return e
}
