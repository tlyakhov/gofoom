// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/controllers"
	"tlyakhov/gofoom/ecs"
	"tlyakhov/gofoom/editor/state"

	"tlyakhov/gofoom/components/core"

	"fyne.io/fyne/v2/driver/desktop"
)

type SplitSector struct {
	state.IEditor

	Splitters []*controllers.SectorSplitter
	Original  map[ecs.Entity][]ecs.Attachable
}

func (a *SplitSector) OnMouseDown(evt *desktop.MouseEvent) {}
func (a *SplitSector) OnMouseMove()                        {}
func (a *SplitSector) Act()                                {}

func (a *SplitSector) Split(sector *core.Sector) {
	s := &controllers.SectorSplitter{
		Splitter1: *a.WorldGrid(&a.State().MouseDownWorld),
		Splitter2: *a.WorldGrid(&a.State().MouseWorld),
		Sector:    sector,
	}
	a.Splitters = append(a.Splitters, s)
	s.Do()
	if s.Result == nil || len(s.Result) == 0 {
		return
	}
	// Copy original sector's components to preserve them
	o := sector.ECS.AllComponents(sector.Entity)
	a.Original[sector.Entity] = make([]ecs.Attachable, len(o))
	copy(a.Original[sector.Entity], o)
	// Detach the original from the ECS
	sector.ECS.DetachAll(sector.Entity)
	// Attach the cloned entities/components
	for _, added := range s.Result {
		for index, component := range added {
			if component == nil {
				continue
			}
			component.GetECS().Attach(index, component.GetEntity(), component)
		}
	}
}

func (a *SplitSector) OnMouseUp() {
	a.Original = make(map[ecs.Entity][]ecs.Attachable)
	a.Splitters = []*controllers.SectorSplitter{}

	var sectors []*core.Sector
	// Split only selected if any, otherwise all sectors.
	if a.State().SelectedObjects.Empty() {
		col := ecs.Column[core.Sector](a.State().ECS, core.SectorComponentIndex)
		sectors = make([]*core.Sector, col.Length)
		for i := range col.Length {
			sectors[i] = col.Value(i)
		}
	} else {
		sectors = make([]*core.Sector, 0)
		visited := make(concepts.Set[ecs.Entity])
		for _, s := range a.State().SelectedObjects.Exact {
			// We could just check for the .Sector field being valid, but then
			// the user may be surprised to have a sector split when they've
			// selected a body or something else.
			if s.Type == core.SelectableEntity || s.Type == core.SelectableBody {
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
	a.ActionFinished(false, true, true)
}

func (a *SplitSector) Cancel() {
	a.ActionFinished(true, true, true)
}

func (a *SplitSector) Undo() {
	bodies := make([]ecs.Entity, 0)

	for _, splitter := range a.Splitters {
		if splitter.Result == nil {
			continue
		}
		for _, addedComponents := range splitter.Result {
			sector := addedComponents[core.SectorComponentIndex].(*core.Sector)
			for entity, body := range sector.Bodies {
				bodies = append(bodies, entity)
				body.SectorEntity = 0
			}
			sector.Bodies = make(map[ecs.Entity]*core.Body)
			a.State().ECS.DetachAll(sector.Entity)
		}
	}
	for entity, originalComponents := range a.Original {
		for index, component := range originalComponents {
			if component == nil {
				continue
			}
			a.State().ECS.Attach(index, entity, component)
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
}
func (a *SplitSector) Redo() {
	bodies := make([]ecs.Entity, 0)

	for entity, originalComponents := range a.Original {
		sector := originalComponents[core.SectorComponentIndex].(*core.Sector)
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
			for index, component := range addedComponents {
				if component == nil {
					continue
				}
				a.State().ECS.Attach(index, component.GetEntity(), component)
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
	}

}

func (a *SplitSector) Status() string {
	return ""
}
