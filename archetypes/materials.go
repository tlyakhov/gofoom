// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package archetypes

import (
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/ecs"
)

func ComponentTableIsMaterial(components ecs.ComponentTable) bool {
	for _, c := range components {
		switch c.(type) {
		case *materials.Shader:
			return true
		case *materials.Sprite:
			return true
		case *materials.Image:
			return true
		case *materials.Text:
			return true
		case *materials.Solid:
			return true
		}
	}
	return false
}

func EntityIsMaterial(db *ecs.ECS, e ecs.Entity) bool {
	return ComponentTableIsMaterial(db.AllComponents(e))
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
