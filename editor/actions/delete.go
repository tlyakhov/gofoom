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

	Selected []any
	Saved    map[uint64]any
}

func (a *Delete) Save(er *concepts.EntityRef) {
	if _, ok := a.Saved[er.Entity]; !ok {
		a.Saved[er.Entity] = a.State().DB.SerializeEntity(er.Entity)
	}
}

func (a *Delete) Act() {
	a.Saved = make(map[uint64]any)
	a.Selected = make([]any, len(a.State().SelectedObjects))
	copy(a.Selected, a.State().SelectedObjects)

	for _, obj := range a.Selected {
		switch target := obj.(type) {
		case *core.SectorSegment:
			a.Save(target.Sector.Ref())
		case *concepts.EntityRef:
			if sector := core.SectorFromDb(target); sector != nil {
				a.Save(target)
			}
			if body := core.BodyFromDb(target); body != nil {
				a.Save(body.SectorEntityRef.Now)
			}
		}
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

	for entity, saved := range a.Saved {
		a.State().DB.DetachAll(entity)
		a.State().DB.DeserializeAndAttachEntity(saved.(map[string]any))
	}

	a.State().DB.ActAllControllers(concepts.ControllerRecalculate)
}
func (a *Delete) Redo() {
	a.State().Lock.Lock()
	defer a.State().Lock.Unlock()

	for _, obj := range a.Selected {
		switch target := obj.(type) {
		case *core.SectorSegment:
			for i, seg := range target.Sector.Segments {
				if seg != target {
					continue
				}
				target.Sector.Segments = append(target.Sector.Segments[:i], target.Sector.Segments[i+1:]...)
				if len(target.Sector.Segments) == 0 {
					target.Sector.Bodies = make(map[uint64]*concepts.EntityRef)
					a.State().DB.DetachAll(target.Sector.Entity)
				}
				break
			}
		case *concepts.EntityRef:
			if behaviors.PlayerFromDb(target) != nil {
				// Otherwise weird things happen...
				continue
			}
			a.State().DB.DetachAll(target.Entity)
			if body := core.BodyFromDb(target); body != nil {
				delete(body.Sector().Bodies, target.Entity)
			}
		}
	}
	a.State().DB.ActAllControllers(concepts.ControllerRecalculate)
}
