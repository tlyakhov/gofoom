// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/editor/state"

	"tlyakhov/gofoom/concepts"

	"fyne.io/fyne/v2/driver/desktop"
)

type Delete struct {
	state.IEditor

	Selected *core.Selection
	Saved    map[*core.Selectable]any
}

func (a *Delete) Act() {
	a.Saved = make(map[*core.Selectable]any)
	a.Selected = core.NewSelectionClone(a.State().SelectedObjects)

	for _, obj := range a.Selected.Exact {
		a.Saved[obj] = obj.Serialize()
	}
	a.Redo()
	a.ActionFinished(false, true, true)
}
func (a *Delete) Cancel()                             {}
func (a *Delete) Frame()                              {}
func (a *Delete) OnMouseDown(evt *desktop.MouseEvent) {}
func (a *Delete) OnMouseMove()                        {}
func (a *Delete) OnMouseUp()                          {}

func (a *Delete) Undo() {
	a.State().Lock.Lock()
	defer a.State().Lock.Unlock()

	for s, saved := range a.Saved {
		s.DB.DetachAll(s.Entity)
		s.DB.DeserializeAndAttachEntity(saved.(map[string]any))
	}

	a.State().DB.ActAllControllers(concepts.ControllerRecalculate)
}
func (a *Delete) Redo() {
	a.State().Lock.Lock()
	defer a.State().Lock.Unlock()

	for s := range a.Saved {
		switch s.Type {
		case core.SelectableSector:
			s.Sector.Bodies = make(map[concepts.Entity]*core.Body)
			s.Sector.InternalSegments = make(map[concepts.Entity]*core.InternalSegment)
			s.DB.DetachAll(s.Sector.Entity)
		case core.SelectableSectorSegment:
			for i, seg := range s.Sector.Segments {
				if seg != s.SectorSegment {
					continue
				}
				s.Sector.Segments = append(s.Sector.Segments[:i], s.Sector.Segments[i+1:]...)
				if len(s.Sector.Segments) == 0 {
					s.Sector.Bodies = make(map[concepts.Entity]*core.Body)
					s.Sector.InternalSegments = make(map[concepts.Entity]*core.InternalSegment)
					s.DB.DetachAll(s.Sector.Entity)
				}
				break
			}
		case core.SelectableBody:
			if behaviors.PlayerFromDb(s.DB, s.Entity) != nil {
				// Otherwise weird things happen...
				continue
			}
			a.State().DB.DetachAll(s.Entity)
			if s.Body.Sector() != nil {
				delete(s.Body.Sector().Bodies, s.Entity)
			}
		case core.SelectableInternalSegment:
			fallthrough
		case core.SelectableInternalSegmentA:
			fallthrough
		case core.SelectableInternalSegmentB:
			s.DB.DetachAll(s.InternalSegment.Entity)
			s.InternalSegment.DetachFromSectors()
		case core.SelectableEntityRef:
			s.DB.DetachAll(s.Entity)
		}
	}
	a.State().DB.ActAllControllers(concepts.ControllerRecalculate)
}
