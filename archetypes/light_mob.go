package archetypes

import (
	"strconv"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
)

func IsLightMob(er concepts.EntityRef) bool {
	return er.Component(core.MobComponentIndex) != nil &&
		er.Component(core.LightComponentIndex) != nil
}

func CreateLightMob(db *concepts.EntityComponentDB) *concepts.EntityRef {
	er := db.NewEntityRef()
	mob := db.NewComponent(er.Entity, core.MobComponentIndex).(*core.Mob)
	named := db.NewComponent(er.Entity, concepts.NamedComponentIndex).(*concepts.Named)
	db.NewComponent(er.Entity, core.LightComponentIndex)
	mob.BoundingRadius = 10.0
	named.Name = "Light " + strconv.FormatUint(er.Entity, 10)

	return er
}
