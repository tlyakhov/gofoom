// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"tlyakhov/gofoom/editor/state"

	"tlyakhov/gofoom/concepts"

	"fyne.io/fyne/v2/driver/desktop"
)

type Move struct {
	state.IEditor

	Selected []*state.Selectable
	Original []concepts.Vector3
	Delta    concepts.Vector2
}

func (a *Move) OnMouseDown(evt *desktop.MouseEvent) {
	a.SetMapCursor(desktop.PointerCursor)

	a.Selected = make([]*state.Selectable, len(a.State().SelectedObjects))
	copy(a.Selected, a.State().SelectedObjects)
	a.Original = make([]concepts.Vector3, len(a.Selected))

	i := 0
	for _, s := range a.Selected {
		a.State().Lock.Lock()
		switch s.Type {
		case state.SelectableSector:
			for _, seg := range s.Sector.Segments {
				a.Original = append(a.Original, concepts.Vector3{})
				seg.P.To3D(&a.Original[i])
				i++
			}
		case state.SelectableSectorSegment:
			a.Original = append(a.Original, concepts.Vector3{})
			s.SectorSegment.P.To3D(&a.Original[i])
			i++
		case state.SelectableBody:
			a.Original = append(a.Original, s.Body.Pos.Original)
			i++
		case state.SelectableInternalSegmentA:
			a.Original = append(a.Original, concepts.Vector3{})
			s.InternalSegment.A.To3D(&a.Original[i])
			i++
		case state.SelectableInternalSegmentB:
			a.Original = append(a.Original, concepts.Vector3{})
			s.InternalSegment.B.To3D(&a.Original[i])
			i++
		case state.SelectableInternalSegment:
			a.Original = append(a.Original, concepts.Vector3{})
			s.InternalSegment.A.To3D(&a.Original[i])
			i++
			a.Original = append(a.Original, concepts.Vector3{})
			s.InternalSegment.B.To3D(&a.Original[i])
			i++
		}
		a.State().Lock.Unlock()
	}
}

func (a *Move) OnMouseMove() {
	a.Delta = *a.State().MouseWorld.Sub(&a.State().MouseDownWorld)
	a.Act()
}

func (a *Move) OnMouseUp() {
	a.State().Modified = true
	a.ActionFinished(false, true, true)
}
func (a *Move) Act() {
	a.State().Lock.Lock()

	i := 0
	for _, s := range a.Selected {
		switch s.Type {
		case state.SelectableSector:
			for _, seg := range s.Sector.Segments {
				seg.P = *a.WorldGrid(a.Original[i].To2D().Add(&a.Delta))
				i++
			}
			s.Sector.Recalculate()
		case state.SelectableSectorSegment:
			s.SectorSegment.P = *a.WorldGrid(a.Original[i].To2D().Add(&a.Delta))
			s.Sector.Recalculate()
			i++
		case state.SelectableBody:
			s.Body.Pos.Original = *a.WorldGrid3D(a.Original[i].Add(a.Delta.To3D(new(concepts.Vector3))))
			s.Body.Pos.Reset()
			i++
		case state.SelectableInternalSegmentA:
			s.InternalSegment.A = a.WorldGrid(a.Original[i].To2D().Add(&a.Delta))
			s.InternalSegment.Recalculate()
			i++
		case state.SelectableInternalSegmentB:
			s.InternalSegment.B = a.WorldGrid(a.Original[i].To2D().Add(&a.Delta))
			s.InternalSegment.Recalculate()
			i++
		case state.SelectableInternalSegment:
			s.InternalSegment.A = a.WorldGrid(a.Original[i].To2D().Add(&a.Delta))
			i++
			s.InternalSegment.B = a.WorldGrid(a.Original[i].To2D().Add(&a.Delta))
			i++
			s.InternalSegment.Recalculate()
		}
	}
	a.State().DB.ActAllControllers(concepts.ControllerRecalculate)
	a.State().Lock.Unlock()
}
func (a *Move) Cancel() {}
func (a *Move) Frame()  {}

func (a *Move) Undo() {
	i := 0
	for _, s := range a.Selected {
		switch s.Type {
		case state.SelectableSector:
			for _, seg := range s.Sector.Segments {
				seg.P = *a.Original[i].To2D()
				i++
			}
			s.Sector.Recalculate()
		case state.SelectableSectorSegment:
			s.SectorSegment.P = *a.Original[i].To2D()
			s.Sector.Recalculate()
			i++
		case state.SelectableBody:
			s.Body.Pos.Original = a.Original[i]
			s.Body.Pos.Reset()
			i++
		case state.SelectableInternalSegmentA:
			s.InternalSegment.A = a.Original[i].To2D()
			i++
			s.InternalSegment.Recalculate()
		case state.SelectableInternalSegmentB:
			s.InternalSegment.B = a.Original[i].To2D()
			i++
			s.InternalSegment.Recalculate()
		case state.SelectableInternalSegment:
			s.InternalSegment.A = a.Original[i].To2D()
			i++
			s.InternalSegment.B = a.Original[i].To2D()
			i++
			s.InternalSegment.Recalculate()
		}
	}
	a.State().DB.ActAllControllers(concepts.ControllerRecalculate)
}
func (a *Move) Redo() {
	a.Act()
}

func (a *Move) RequiresLock() bool { return true }
