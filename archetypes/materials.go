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
		case *materials.SpriteSheet:
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

func EntityIsMaterial(u *ecs.Universe, e ecs.Entity) bool {
	return ComponentTableIsMaterial(u.AllComponents(e))
}

func CreateBasicMaterial(u *ecs.Universe, textured bool) ecs.Entity {
	e := u.NewEntity()
	named := u.NewAttachedComponent(e, ecs.NamedCID).(*ecs.Named)
	named.Name = "Material " + e.Format(u)
	u.NewAttachedComponent(e, materials.LitCID)
	if textured {
		u.NewAttachedComponent(e, materials.ImageCID)
	} else {
		u.NewAttachedComponent(e, materials.SolidCID)
	}
	return e
}
