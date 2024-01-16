package archetypes

import (
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
)

func IsPlayerBody(er concepts.EntityRef) bool {
	return er.Component(core.BodyComponentIndex) != nil &&
		er.Component(behaviors.PlayerComponentIndex) != nil &&
		er.Component(behaviors.AliveComponentIndex) != nil
}

func CreatePlayerBody(db *concepts.EntityComponentDB) *concepts.EntityRef {
	er := db.NewEntityRef()
	body := db.NewComponent(er.Entity, core.BodyComponentIndex).(*core.Body)
	_ = db.NewComponent(er.Entity, behaviors.PlayerComponentIndex).(*behaviors.Player)
	_ = db.NewComponent(er.Entity, behaviors.AliveComponentIndex).(*behaviors.Alive)

	body.Size.Set(concepts.Vector2{15, constants.PlayerHeight})
	body.BoundingRadius = constants.PlayerBoundingRadius
	body.Mass = constants.PlayerMass // kg

	return er
}
