package archetypes

import (
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
)

func IsPlayerMob(er concepts.EntityRef) bool {
	return er.Component(core.MobComponentIndex) != nil &&
		er.Component(behaviors.PlayerComponentIndex) != nil &&
		er.Component(behaviors.AliveComponentIndex) != nil
}

func CreatePlayerMob(db *concepts.EntityComponentDB) *concepts.EntityRef {
	er := db.NewEntityRef()
	mob := db.NewComponent(er.Entity, core.MobComponentIndex).(*core.Mob)
	_ = db.NewComponent(er.Entity, behaviors.PlayerComponentIndex).(*behaviors.Player)
	_ = db.NewComponent(er.Entity, behaviors.AliveComponentIndex).(*behaviors.Alive)

	mob.Height = constants.PlayerHeight
	mob.BoundingRadius = constants.PlayerBoundingRadius
	mob.Mass = constants.PlayerMass // kg

	return er
}
