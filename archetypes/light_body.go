package archetypes

import (
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
	db.NewComponent(er.Entity, core.LightComponentIndex)
	body.Size.Original[0] = 2
	body.Size.Original[1] = 2
	body.Size.Reset()

	return er
}
