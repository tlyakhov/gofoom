// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/editor/state"

	"tlyakhov/gofoom/concepts"

	"fyne.io/fyne/v2/driver/desktop"
)

type Delete struct {
	state.IEditor

	Selected []*state.Selectable
	Saved    map[*state.Selectable]any
}

func (a *Delete) Act() {
	a.Saved = make(map[*state.Selectable]any)
	a.Selected = make([]*state.Selectable, len(a.State().SelectedObjects))
	copy(a.Selected, a.State().SelectedObjects)

	for _, obj := range a.Selected {
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
		switch s.Type {
		case state.SelectableSector:
			fallthrough
		case state.SelectableSectorSegment:
			// Reattach the whole sector, in case the user deleted the last segment
			a.State().DB.DetachAll(s.Sector.Entity)
		case state.SelectableBody:
			a.State().DB.DetachAll(s.Body.Entity)
		case state.SelectableInternalSegment:
			fallthrough
		case state.SelectableInternalSegmentA:
			fallthrough
		case state.SelectableInternalSegmentB:
			a.State().DB.DetachAll(s.InternalSegment.Entity)
		}
		a.State().DB.DeserializeAndAttachEntity(saved.(map[string]any))
	}

	a.State().DB.ActAllControllers(concepts.ControllerRecalculate)
}
func (a *Delete) Redo() {
	a.State().Lock.Lock()
	defer a.State().Lock.Unlock()

	for s := range a.Saved {
		switch s.Type {
		case state.SelectableSector:
			s.Sector.Bodies = make(map[uint64]*concepts.EntityRef)
			s.Sector.InternalSegments = make(map[uint64]*concepts.EntityRef)
			a.State().DB.DetachAll(s.Sector.Entity)
		case state.SelectableSectorSegment:
			for i, seg := range s.Sector.Segments {
				if seg != s.SectorSegment {
					continue
				}
				s.Sector.Segments = append(s.Sector.Segments[:i], s.Sector.Segments[i+1:]...)
				if len(s.Sector.Segments) == 0 {
					s.Sector.Bodies = make(map[uint64]*concepts.EntityRef)
					s.Sector.InternalSegments = make(map[uint64]*concepts.EntityRef)
					a.State().DB.DetachAll(s.Sector.Entity)
				}
				break
			}
		case state.SelectableBody:
			if behaviors.PlayerFromDb(s.Body.EntityRef) != nil {
				// Otherwise weird things happen...
				continue
			}
			a.State().DB.DetachAll(s.Body.Entity)
			if s.Body.Sector() != nil {
				delete(s.Body.Sector().Bodies, s.Body.Entity)
			}
		case state.SelectableInternalSegment:
			fallthrough
		case state.SelectableInternalSegmentA:
			fallthrough
		case state.SelectableInternalSegmentB:
			a.State().DB.DetachAll(s.InternalSegment.Entity)
			if s.InternalSegment.Sector() != nil {
				delete(s.InternalSegment.Sector().Bodies, s.InternalSegment.Entity)
			}
		}
	}
	a.State().DB.ActAllControllers(concepts.ControllerRecalculate)
}
