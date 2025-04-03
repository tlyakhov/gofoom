// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
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
	Saved    map[*selection.Selectable]any
}

func (a *Delete) Activate() {
	a.Saved = make(map[*selection.Selectable]any)
	a.Selected = selection.NewSelectionClone(a.State().SelectedObjects)

	for _, obj := range a.Selected.Exact {
		a.Saved[obj] = obj.Serialize()
	}
	a.apply()
	a.ActionFinished(false, true, true)
}

func (a *Delete) Undo() {
	u := a.State().Universe
	for s, saved := range a.Saved {
		data := saved.(map[string]any)
		for name, cid := range ecs.Types().IDs {
			yamlData := data[name]
			if yamlData == nil {
				continue
			}

			if yamlLink, ok := yamlData.(string); ok {
				linkedEntity, _ := ecs.ParseEntity(yamlLink)
				if linkedEntity != 0 {
					c := u.Component(linkedEntity, cid)
					if c != nil {
						u.Attach(cid, s.Entity, &c)
					}
				}
			} else {
				yamlComponent := yamlData.(map[string]any)
				var attached ecs.Attachable
				u.Attach(cid, s.Entity, &attached)
				if attached.Base().Attachments == 1 {
					attached.Construct(yamlComponent)
				}
			}
		}
	}

	a.State().Universe.ActAllControllers(ecs.ControllerRecalculate)
}
func (a *Delete) Redo() {
	a.apply()
}

func (a *Delete) apply() {
	for s := range a.Saved {
		switch s.Type {
		case selection.SelectableSector:
			if s.Sector.IsExternal() {
				continue
			}
			s.Sector.Bodies = make(map[ecs.Entity]*core.Body)
			s.Sector.InternalSegments = make(map[ecs.Entity]*core.InternalSegment)
			s.Universe.Delete(s.Sector.Entity)
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
					s.Universe.Delete(s.Sector.Entity)
				}
				break
			}
		case selection.SelectableBody:
			if s.Body.IsExternal() {
				continue
			}
			if p := character.GetPlayer(s.Universe, s.Entity); p != nil && !p.Spawn {
				// Otherwise weird things happen...
				continue
			}
			a.State().Universe.Delete(s.Entity)
		case selection.SelectableInternalSegment, selection.SelectableInternalSegmentA, selection.SelectableInternalSegmentB:
			if s.InternalSegment.IsExternal() {
				continue
			}
			s.Universe.Delete(s.InternalSegment.Entity)
		case selection.SelectableEntity:
			if s.Entity.IsExternal() {
				continue
			}
			s.Universe.Delete(s.Entity)
		}
	}
	a.State().Universe.ActAllControllers(ecs.ControllerRecalculate)
}
