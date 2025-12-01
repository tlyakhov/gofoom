// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/character"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/selection"
	"tlyakhov/gofoom/ecs"
	"tlyakhov/gofoom/editor/state"
)

// TODO: Prevent deletion of entities from SourceFiles
type Delete struct {
	state.Action

	Selected *selection.Selection
}

func (a *Delete) Activate() {
	a.Selected = selection.NewSelectionClone(a.State().Selection)
	a.apply()
	a.State().Modified = true
	a.ActionFinished(false, true, true)
}

func (a *Delete) apply() {
	for _, s := range a.Selected.Exact {
		switch s.Type {
		case selection.SelectableSector:
			if s.Sector.IsExternal() {
				continue
			}
			s.Sector.Bodies = make(map[ecs.Entity]*core.Body)
			s.Sector.InternalSegments = make(map[ecs.Entity]*core.InternalSegment)
			ecs.Delete(s.Sector.Entity)
		case selection.SelectableSectorSegment:
			if s.Sector.IsExternal() {
				continue
			}
			for i, seg := range s.Sector.Segments {
				if seg != s.SectorSegment {
					continue
				}
				s.Sector.Segments = append(s.Sector.Segments[:i], s.Sector.Segments[i+1:]...)
				if len(s.Sector.Segments) == 0 {
					s.Sector.Bodies = make(map[ecs.Entity]*core.Body)
					s.Sector.InternalSegments = make(map[ecs.Entity]*core.InternalSegment)
					ecs.Delete(s.Sector.Entity)
				}
				break
			}
		case selection.SelectableBody:
			if s.Body.IsExternal() {
				continue
			}
			if p := character.GetPlayer(s.Entity); p != nil {
				if s := behaviors.GetSpawner(s.Entity); s == nil {
					continue
				}
			}
			ecs.Delete(s.Entity)
		case selection.SelectableInternalSegment, selection.SelectableInternalSegmentA, selection.SelectableInternalSegmentB:
			if s.InternalSegment.IsExternal() {
				continue
			}
			ecs.Delete(s.InternalSegment.Entity)
		case selection.SelectableEntity:
			if s.Entity.IsExternal() {
				continue
			}
			ecs.Delete(s.Entity)
		}
	}
	ecs.ActAllControllers(ecs.ControllerRecalculate)
}
