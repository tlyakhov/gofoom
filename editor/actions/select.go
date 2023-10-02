package actions

import (
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/editor/state"

	"github.com/gotk3/gotk3/gdk"
)

type SelectModifier int

const (
	SelectNew SelectModifier = iota
	SelectAdd
	SelectSub
)

type Select struct {
	state.IEditor

	Mode     string
	Modifier SelectModifier
	Original []concepts.ISerializable
	Selected []concepts.ISerializable
}

func (a *Select) OnMouseDown(button *gdk.EventButton) {
	if button.State()&uint(gdk.SHIFT_MASK) != 0 {
		a.Modifier = SelectAdd
	} else if button.State()&uint(gdk.META_MASK) != 0 {
		a.Modifier = SelectSub
	}

	a.Original = make([]concepts.ISerializable, len(a.State().SelectedObjects))
	for i, o := range a.State().SelectedObjects {
		a.Original[i] = o
	}

	a.Mode = "SelectionStart"
	a.SetMapCursor("cell")
}

func (a *Select) OnMouseMove() {
	a.Mode = "Selecting"
}

func (a *Select) OnMouseUp() {
	hovering := a.State().HoveringObjects

	if len(hovering) == 0 { // User is trying to select a sector?
		hovering = []concepts.ISerializable{}
		for _, sector := range a.State().World.Sectors {
			if sector.IsPointInside2D(a.State().MouseWorld) {
				hovering = append(hovering, sector)
			}
		}
	}

	if a.Modifier == SelectAdd {
		a.Selected = make([]concepts.ISerializable, len(a.Original))
		copy(a.Selected, a.Original)
		a.Selected = append(a.Selected, hovering...)
	} else if a.Modifier == SelectSub {
		a.Selected = []concepts.ISerializable{}
		for _, obj := range a.Original {
			if concepts.IndexOf(hovering, obj) == -1 {
				a.Selected = append(a.Selected, obj)
			}
		}
	} else {
		a.Selected = make([]concepts.ISerializable, len(hovering))
		copy(a.Selected, hovering)
	}
	a.SelectObjects(a.Selected)
	a.ActionFinished(false)
}
func (a *Select) Act()    {}
func (a *Select) Cancel() {}
func (a *Select) Frame()  {}

func (a *Select) Undo() {
	a.SelectObjects(a.Original)
}
func (a *Select) Redo() {
	a.SelectObjects(a.Selected)
}
