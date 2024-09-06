// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/selection"
	"tlyakhov/gofoom/ecs"
	"tlyakhov/gofoom/editor/state"
)

type Delete struct {
	state.IEditor

	Selected *selection.Selection
	Saved    map[*selection.Selectable]any
}

func (a *Delete) Act() {
	a.Saved = make(map[*selection.Selectable]any)
	a.Selected = selection.NewSelectionClone(a.State().SelectedObjects)

	for _, obj := range a.Selected.Exact {
		a.Saved[obj] = obj.Serialize()
	}
	a.Redo()
	a.ActionFinished(false, true, true)
}

func (a *Delete) Undo() {
	a.State().Lock.Lock()
	defer a.State().Lock.Unlock()

	for s, saved := range a.Saved {
		s.ECS.DetachAll(s.Entity)
		s.ECS.DeserializeAndAttachEntity(saved.(map[string]any))
	}

	a.State().ECS.ActAllControllers(ecs.ControllerRecalculate)
}
func (a *Delete) Redo() {
	a.State().Lock.Lock()
	defer a.State().Lock.Unlock()

	for s := range a.Saved {
		switch s.Type {
		case selection.SelectableSector:
			s.Sector.Bodies = make(map[ecs.Entity]*core.Body)
			s.Sector.InternalSegments = make(map[ecs.Entity]*core.InternalSegment)
			s.ECS.DetachAll(s.Sector.Entity)
		case selection.SelectableSectorSegment:
			for i, seg := range s.Sector.Segments {
				if seg != s.SectorSegment {
					continue
				}
				s.Sector.Segments = append(s.Sector.Segments[:i], s.Sector.Segments[i+1:]...)
				if len(s.Sector.Segments) == 0 {
					s.Sector.Bodies = make(map[ecs.Entity]*core.Body)
					s.Sector.InternalSegments = make(map[ecs.Entity]*core.InternalSegment)
					s.ECS.DetachAll(s.Sector.Entity)
				}
				break
			}
		case selection.SelectableBody:
			if behaviors.GetPlayer(s.ECS, s.Entity) != nil {
				// Otherwise weird things happen...
				continue
			}
			a.State().ECS.DetachAll(s.Entity)
			if s.Body.Sector() != nil {
				delete(s.Body.Sector().Bodies, s.Entity)
			}
		case selection.SelectableInternalSegment:
			fallthrough
		case selection.SelectableInternalSegmentA:
			fallthrough
		case selection.SelectableInternalSegmentB:
			s.ECS.DetachAll(s.InternalSegment.Entity)
			s.InternalSegment.DetachFromSectors()
		case selection.SelectableEntity:
			s.ECS.DetachAll(s.Entity)
		}
	}
	a.State().ECS.ActAllControllers(ecs.ControllerRecalculate)
}
