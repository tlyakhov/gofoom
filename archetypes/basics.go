package archetypes

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
)

func CreateBasic(db *concepts.EntityComponentDB, componentIndex int) *concepts.EntityRef {
	er := db.NewEntityRef()
	db.NewComponent(er.Entity, componentIndex)

	return er
}

func CreateSector(db *concepts.EntityComponentDB) *concepts.EntityRef {
	er := db.NewEntityRef()
	db.NewComponent(er.Entity, core.SectorComponentIndex)

	return er
}
