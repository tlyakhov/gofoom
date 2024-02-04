package archetypes

import (
	"strconv"
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

func EntityRefIsMaterial(er *concepts.EntityRef) bool {
	return EntityMapIsMaterial(er.All())
}

func AttachableIsMaterial(a concepts.Attachable) bool {
	return EntityMapIsMaterial(a.Ref().All())
}

func CreateBasicMaterial(db *concepts.EntityComponentDB, textured bool) *concepts.EntityRef {
	er := db.RefForNewEntity()
	named := db.NewAttachedComponent(er.Entity, concepts.NamedComponentIndex).(*concepts.Named)
	named.Name = "Material " + strconv.FormatUint(er.Entity, 10)
	db.NewAttachedComponent(er.Entity, materials.LitComponentIndex)
	if textured {
		db.NewAttachedComponent(er.Entity, materials.ImageComponentIndex)
	} else {
		db.NewAttachedComponent(er.Entity, materials.SolidComponentIndex)
	}
	return er
}
