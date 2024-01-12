package archetypes

import (
	"strconv"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"
)

func EntityMapIsMaterial(components []concepts.Attachable) bool {
	return components[materials.ShaderComponentIndex] != nil ||
		components[materials.LitComponentIndex] != nil ||
		components[materials.TiledComponentIndex] != nil ||
		components[materials.SkyComponentIndex] != nil ||
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
	er := db.NewEntityRef()
	named := db.NewComponent(er.Entity, concepts.NamedComponentIndex).(*concepts.Named)
	named.Name = "Material " + strconv.FormatUint(er.Entity, 10)
	db.NewComponent(er.Entity, materials.LitComponentIndex)
	if textured {
		db.NewComponent(er.Entity, materials.TiledComponentIndex)
		db.NewComponent(er.Entity, materials.ImageComponentIndex)
	} else {
		db.NewComponent(er.Entity, materials.SolidComponentIndex)
	}
	return er
}
