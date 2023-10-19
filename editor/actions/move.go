package actions

import (
	"tlyakhov/gofoom/editor/state"
	"tlyakhov/gofoom/logic/provide"

	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/core"

	"github.com/gotk3/gotk3/gdk"
)

type Move struct {
	state.IEditor

	Selected []concepts.ISerializable
	Original []concepts.Vector3
	Delta    concepts.Vector2
}

func (a *Move) OnMouseDown(button *gdk.EventButton) {
	a.SetMapCursor("move")

	a.Selected = []concepts.ISerializable{}
	for _, obj := range a.State().SelectedObjects {
		if sector, ok := obj.(core.AbstractSector); ok {
			for _, seg := range sector.Physical().Segments {
				a.Selected = append(a.Selected, &state.MapPoint{Segment: seg})
			}
		} else {
			a.Selected = append(a.Selected, obj)
		}
	}

	a.Original = make([]concepts.Vector3, len(a.Selected))
	for i, obj := range a.Selected {
		switch target := obj.(type) {
		case *state.MapPoint:
			target.P.To3D(&a.Original[i])
		case core.AbstractEntity:
			a.Original[i] = target.Physical().Pos.Original
		}
	}
}

func (a *Move) OnMouseMove() {
	a.Delta = *a.State().MouseWorld.Sub(&a.State().MouseDownWorld)
	a.Act()
}

func (a *Move) OnMouseUp() {
	a.SelectObjects(a.State().SelectedObjects) // Updates properties.
	a.State().Modified = true
	a.ActionFinished(false)
}
func (a *Move) Act() {
	for i, obj := range a.Selected {
		switch target := obj.(type) {
		case *state.MapPoint:
			target.P = *a.WorldGrid(a.Original[i].To2D().Add(&a.Delta))
			provide.Passer.For(target.Sector).Recalculate()
		case core.AbstractEntity:
			if target == a.State().World.Player {
				// Otherwise weird things happen...
				continue
			}
			target.Physical().Pos.Original = *a.WorldGrid3D(a.Original[i].Add(a.Delta.To3D(new(concepts.Vector3))))
			target.Physical().Pos.Reset()
			a.State().World.Recalculate()
		}
	}
}
func (a *Move) Cancel() {}
func (a *Move) Frame()  {}

func (a *Move) Undo() {
	for i, obj := range a.Selected {
		switch target := obj.(type) {
		case *state.MapPoint:
			target.P = *a.Original[i].To2D()
			provide.Passer.For(target.Sector).Recalculate()
		case core.AbstractEntity:
			if target == a.State().World.Player {
				// Otherwise weird things happen...
				continue
			}
			target.Physical().Pos.Original = a.Original[i]
			target.Physical().Pos.Reset()
			a.State().World.Recalculate()
		}
	}
}
func (a *Move) Redo() {
	a.Act()
}
