package archetypes

import (
	"strconv"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
)

func CreateBasic(db *concepts.EntityComponentDB, componentIndex int) *concepts.EntityRef {
	er := db.NewEntityRef()
	named := db.NewComponent(er.Entity, concepts.NamedComponentIndex).(*concepts.Named)
	named.Name = "Basic " + strconv.FormatUint(er.Entity, 10)
	db.NewComponent(er.Entity, componentIndex)

	return er
}

func CreateSector(db *concepts.EntityComponentDB) *concepts.EntityRef {
	er := db.NewEntityRef()
	named := db.NewComponent(er.Entity, concepts.NamedComponentIndex).(*concepts.Named)
	named.Name = "Sector " + strconv.FormatUint(er.Entity, 10)
	db.NewComponent(er.Entity, core.SectorComponentIndex)

	return er
}
