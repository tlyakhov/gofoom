package main

import (
	"github.com/gotk3/gotk3/gdk"
	"github.com/tlyakhov/gofoom/concepts"
)

type SelectMode int

const (
	SelectNew SelectMode = iota
	SelectAdd
	SelectSub
)

type SelectAction struct {
	*Editor
	Mode     SelectMode
	Original []concepts.ISerializable
	Selected []concepts.ISerializable
}

func (a *SelectAction) OnMouseDown(button *gdk.EventButton) {
	if button.State()&uint(gdk.GDK_SHIFT_MASK) != 0 {
		a.Mode = SelectAdd
	} else if button.State()&uint(gdk.GDK_META_MASK) != 0 {
		a.Mode = SelectSub
	}

	a.State = "SelectionStart"
	a.SetMapCursor("cell")
}

func (a *SelectAction) OnMouseMove() {
	a.State = "Selecting"
}

func (a *SelectAction) OnMouseUp() {
	hovering := a.HoveringObjects

	if len(hovering) == 0 { // User is trying to select a sector?
		hovering = []concepts.ISerializable{}
		for _, sector := range a.GameMap.Sectors {
			if sector.IsPointInside2D(a.MouseWorld) {
				hovering = append(hovering, sector)
			}
		}
	}

	if a.Mode == SelectAdd {
		a.Selected = make([]concepts.ISerializable, len(a.Original))
		copy(a.Selected, a.Original)
		a.Selected = append(a.Selected, hovering...)
	} else if a.Mode == SelectSub {
		a.Selected = []concepts.ISerializable{}
		for _, obj := range a.Original {
			if indexOfObject(hovering, obj) == -1 {
				a.Selected = append(a.Selected, obj)
			}
		}
	} else {
		a.Selected = make([]concepts.ISerializable, len(hovering))
		copy(a.Selected, hovering)
	}
	a.SelectObjects(a.Selected)
	a.ActionFinished()
}
func (a *SelectAction) Act()    {}
func (a *SelectAction) Cancel() {}
func (a *SelectAction) Frame()  {}

func (a *SelectAction) Undo() {
	a.SelectObjects(a.Original)
}
func (a *SelectAction) Redo() {
	a.SelectObjects(a.Selected)
}
