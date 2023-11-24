package controllers

import (
	"tlyakhov/gofoom/archetypes"
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"
)

type WorldController struct {
	concepts.BaseController
	*core.Spawn
}

func init() {
	concepts.DbTypes().RegisterController(&WorldController{})
}

func (wc *WorldController) Priority() int {
	return 70
}

func (wc *WorldController) Methods() concepts.ControllerMethod {
	return concepts.ControllerLoaded | concepts.ControllerAlways
}

func (wc *WorldController) Target(target *concepts.EntityRef) bool {
	wc.TargetEntity = target
	wc.Spawn = core.SpawnFromDb(target)
	return wc.Spawn != nil && wc.Active
}

func (wc *WorldController) Loaded() {
	// Create a player if we don't have one
	if wc.DB.First(behaviors.PlayerComponentIndex) == nil {
		player := archetypes.CreatePlayerBody(wc.DB)
		playerBody := core.BodyFromDb(player)
		playerBody.Pos.Original = wc.Spawn.Spawn
		playerBody.Pos.Reset()
	}
}

func (wc *WorldController) Always() {
	set := wc.NewControllerSet()
	for _, c := range wc.DB.Components[core.BodyComponentIndex] {
		body := c.(*core.Body)
		if !body.Active || body.SectorEntityRef.Nil() {
			continue
		}
		for _, pvs := range body.Sector().PVSBody {
			// First, consider the case where the sector entity has a proximity
			// component that includes the body as a valid triggering source
			if proximity := behaviors.ProximityFromDb(pvs.Ref()); proximity != nil && proximity.Active {
				if proximity.Condition.Valid(body.Ref()) &&
					pvs.Center.Dist2(&body.Pos.Now) < proximity.Range*proximity.Range {
					set.Act(pvs.Ref(), body.Ref(), concepts.ControllerTrigger)
				}
			}

			// Next, consider the case where the body entity has a proximity
			// component that includes the sector as a valid triggering source
			if proximity := behaviors.ProximityFromDb(body.Ref()); proximity != nil && proximity.Active {
				if proximity.Condition.Valid(pvs.Ref()) &&
					pvs.Center.Dist2(&body.Pos.Now) < proximity.Range*proximity.Range {
					set.Act(body.Ref(), pvs.Ref(), concepts.ControllerTrigger)
				}
			}
		}
	}
}

func DefaultMaterial(db *concepts.EntityComponentDB) *concepts.EntityRef {
	er := db.GetEntityRefByName("Default Material")
	if !er.Nil() {
		return er
	}

	// Otherwise try a random one?
	return db.First(materials.LitComponentIndex).Ref()
}
