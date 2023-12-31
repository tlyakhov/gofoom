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

func (wc *WorldController) proximity(sector *core.Sector, body *core.Body, set *concepts.ControllerSet) {
	// Consider the case where the sector entity has a proximity
	// component that includes the body as a valid scripting source
	if p := behaviors.ProximityFromDb(sector.Ref()); p != nil && p.Active {
		if sector.Center.Dist2(&body.Pos.Now) < p.Range*p.Range {
			BodySectorScript(p.Scripts, body.Ref(), sector.Ref())
		}
	}

	// Consider the case where the body entity has a proximity
	// component that includes the sector as a valid scripting source
	if p := behaviors.ProximityFromDb(body.Ref()); p != nil && p.Active {
		if sector.Center.Dist2(&body.Pos.Now) < p.Range*p.Range {
			BodySectorScript(p.Scripts, body.Ref(), sector.Ref())
		}
	}
}

func (wc *WorldController) Always() {
	set := wc.NewControllerSet()
	for _, c := range wc.DB.Components[core.BodyComponentIndex] {
		body := c.(*core.Body)
		if !body.Active || body.SectorEntityRef.Nil() {
			continue
		}
		for _, pvs := range body.Sector().PVS {
			wc.proximity(pvs, body, set)
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
