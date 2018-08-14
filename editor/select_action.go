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
	Original map[string]concepts.ISerializable
	Selected map[string]concepts.ISerializable
}

func (a *SelectAction) OnMouseDown(button *gdk.EventButton) {
	if button.State()&uint(gdk.GDK_SHIFT_MASK) != 0 {
		a.Mode = SelectAdd
	} else if button.State()&uint(gdk.GDK_META_MASK) != 0 {
		a.Mode = SelectSub
	}

	a.State = "SelectionStart"
	// SetCursor
}

func (a *SelectAction) OnMouseMove() {
	a.State = "Selecting"
}

func (a *SelectAction) OnMouseUp() {
	hovering := a.HoveringObjects

	if hovering == nil || len(hovering) == 0 { // User is trying to select a sector?
		hovering = make(map[string]concepts.ISerializable)
		for id, sector := range a.GameMap.Sectors {
			if sector.IsPointInside2D(a.MouseWorld) {
				hovering[id] = sector
			}
		}
	}

	if a.Mode == SelectAdd {
		a.Selected = make(map[string]concepts.ISerializable)
	}
	a.ActionFinished()
}
func (a *SelectAction) Act()    {}
func (a *SelectAction) Cancel() {}
func (a *SelectAction) Frame()  {}

func (a *SelectAction) Undo() {
}
func (a *SelectAction) Redo() {
}
