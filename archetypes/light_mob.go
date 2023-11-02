package archetypes

import (
	"strconv"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
)

func IsLightBody(er concepts.EntityRef) bool {
	return er.Component(core.BodyComponentIndex) != nil &&
		er.Component(core.LightComponentIndex) != nil
}

func CreateLightBody(db *concepts.EntityComponentDB) *concepts.EntityRef {
	er := db.NewEntityRef()
	body := db.NewComponent(er.Entity, core.BodyComponentIndex).(*core.Body)
	named := db.NewComponent(er.Entity, concepts.NamedComponentIndex).(*concepts.Named)
	db.NewComponent(er.Entity, core.LightComponentIndex)
	body.BoundingRadius = 10.0
	named.Name = "Light " + strconv.FormatUint(er.Entity, 10)

	return er
}
