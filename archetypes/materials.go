// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package archetypes

import (
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"
)

func EntityMapIsMaterial(components []concepts.Attachable) bool {
	return components[materials.ShaderComponentIndex] != nil ||
		components[materials.LitComponentIndex] != nil ||
		components[materials.ImageComponentIndex] != nil ||
		components[materials.TextComponentIndex] != nil ||
		components[materials.SolidComponentIndex] != nil
}

func EntityIsMaterial(db *concepts.EntityComponentDB, e concepts.Entity) bool {
	return EntityMapIsMaterial(db.AllComponents(e))
}

func AttachableIsMaterial(a concepts.Attachable) bool {
	return EntityIsMaterial(a.GetDB(), a.GetEntity())
}

func CreateBasicMaterial(db *concepts.EntityComponentDB, textured bool) concepts.Entity {
	e := db.NewEntity()
	named := db.NewAttachedComponent(e, concepts.NamedComponentIndex).(*concepts.Named)
	named.Name = "Material " + e.String(db)
	db.NewAttachedComponent(e, materials.LitComponentIndex)
	if textured {
		db.NewAttachedComponent(e, materials.ImageComponentIndex)
	} else {
		db.NewAttachedComponent(e, materials.SolidComponentIndex)
	}
	return e
}
