// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"tlyakhov/gofoom/containers"
	"tlyakhov/gofoom/controllers"
	"tlyakhov/gofoom/ecs"

	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/selection"
)

type SplitSector struct {
	Place

	Splitters []*controllers.SectorSplitter
	Original  map[ecs.Entity]map[string]any
}

func (a *SplitSector) Act() {}

func (a *SplitSector) Split(sector *core.Sector) {
	s := &controllers.SectorSplitter{
		Splitter1: *a.WorldGrid(&a.State().MouseDownWorld),
		Splitter2: *a.WorldGrid(&a.State().MouseWorld),
		Sector:    sector,
	}
	a.Splitters = append(a.Splitters, s)
	s.Do()
	if len(s.Result) == 0 {
		return
	}
	db := a.State().ECS
	// Copy original sector's components to preserve them
	a.Original[sector.Entity] = db.SerializeEntity(sector.Entity)
	// Detach the original from the ECS
	db.Delete(sector.Entity)
	// Attach the cloned entities/components
	for _, added := range s.Result {
		entity := db.NewEntity()
		for _, component := range added {
			db.Attach(ecs.Types().ID(component), entity, component)
		}
	}
}

func (a *SplitSector) EndPoint() bool {
	if !a.Place.EndPoint() {
		return false
	}

	a.State().Lock.Lock()

	a.Original = make(map[ecs.Entity]map[string]any)
	a.Splitters = []*controllers.SectorSplitter{}

	var sectors []*core.Sector
	// Split only selected if any, otherwise all sectors.
	if a.State().SelectedObjects.Empty() {
		col := ecs.ColumnFor[core.Sector](a.State().ECS, core.SectorCID)
		sectors = make([]*core.Sector, 0)
		for i := range col.Cap() {
			if sector := col.Value(i); sector != nil {
				sectors = append(sectors, sector)
			}
		}
	} else {
		sectors = make([]*core.Sector, 0)
		visited := make(containers.Set[ecs.Entity])
		for _, s := range a.State().SelectedObjects.Exact {
			// We could just check for the .Sector field being valid, but then
			// the user may be surprised to have a sector split when they've
			// selected a body or something else.
			if s.Type == selection.SelectableEntity || s.Type == selection.SelectableBody {
				continue
			}
			if visited.Contains(s.Sector.Entity) {
				continue
			}
			sectors = append(sectors, s.Sector)
			visited.Add(s.Sector.Entity)
		}
	}

	for _, s := range sectors {
		a.Split(s)
	}
	a.State().Modified = true
	a.State().Lock.Unlock()
	a.ActionFinished(false, true, true)
	return true
}

func (a *SplitSector) Cancel() {
	a.ActionFinished(true, true, true)
}

func (a *SplitSector) Undo() {
	/*bodies := make([]ecs.Entity, 0)

	for _, splitter := range a.Splitters {
		if splitter.Result == nil {
			continue
		}
		for _, addedComponents := range splitter.Result {
			sector := addedComponents[core.SectorCID].(*core.Sector)
			for entity, body := range sector.Bodies {
				bodies = append(bodies, entity)
				body.SectorEntity = 0
			}
			sector.Bodies = make(map[ecs.Entity]*core.Body)
			a.State().ECS.DetachAll(sector.Entity)
		}
	}
	for entity, originalComponents := range a.Original {
		for _, component := range originalComponents {
			if component == nil {
				continue
			}
			a.State().ECS.Attach(ecs.Types().ID(component), entity, component)
			if sector, ok := component.(*core.Sector); ok {
				for _, entity := range bodies {
					if body := core.GetBody(a.State().ECS, entity); body != nil {
						if sector.IsPointInside2D(body.Pos.Original.To2D()) {
							body.SectorEntity = sector.Entity
							sector.Bodies[entity] = body
						}
					}
				}
			}
		}
	}*/
}
func (a *SplitSector) Redo() {
	/*	bodies := make([]ecs.Entity, 0)

		for entity, originalComponents := range a.Original {
			sector := originalComponents[core.SectorCID].(*core.Sector)
			for entity, body := range sector.Bodies {
				bodies = append(bodies, entity)
				body.SectorEntity = 0
			}
			sector.Bodies = make(map[ecs.Entity]*core.Body)
			a.State().ECS.DetachAll(entity)
		}

		for _, splitter := range a.Splitters {
			if splitter.Result == nil {
				continue
			}
			for _, addedComponents := range splitter.Result {
				for _, component := range addedComponents {
					if component == nil {
						continue
					}
					a.State().ECS.Attach(ecs.Types().ID(component), component.GetEntity(), component)
					if sector, ok := component.(*core.Sector); ok {
						for _, entity := range bodies {
							if body := core.GetBody(a.State().ECS, entity); body != nil {
								if sector.IsPointInside2D(body.Pos.Original.To2D()) {
									body.SectorEntity = sector.Entity
									sector.Bodies[entity] = body
								}
							}
						}
					}
				}
			}
		}*/

}

func (a *SplitSector) Status() string {
	return ""
}
