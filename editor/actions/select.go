// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/editor/state"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
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
	Original []*core.Selectable
	Selected []*core.Selectable
}

func (a *Select) OnMouseDown(evt *desktop.MouseEvent) {

	if evt.Modifier&fyne.KeyModifierShift != 0 {
		a.Modifier = SelectAdd
	} else if evt.Modifier&fyne.KeyModifierSuper != 0 {
		a.Modifier = SelectSub
	}

	a.Original = make([]*core.Selectable, len(a.State().SelectedObjects))
	copy(a.Original, a.State().SelectedObjects)

	a.Mode = "SelectionStart"
	a.SetMapCursor(desktop.TextCursor)
}

func (a *Select) OnMouseMove() {
	a.Mode = "Selecting"
}

func (a *Select) OnMouseUp() {
	hovering := a.State().HoveringObjects
	if len(hovering) == 0 { // User is trying to select a sector?
		hovering = []*core.Selectable{}
		for _, isector := range a.State().DB.All(core.SectorComponentIndex) {
			sector := isector.(*core.Sector)
			if sector.IsPointInside2D(&a.State().MouseWorld) {
				hovering = append(hovering, core.SelectableFromSector(sector))
			}
		}
	}

	if a.Modifier == SelectAdd {
		a.Selected = make([]*core.Selectable, len(a.Original))
		copy(a.Selected, a.Original)
		a.Selected = append(a.Selected, hovering...)
	} else if a.Modifier == SelectSub {
		a.Selected = []*core.Selectable{}
		for _, obj := range a.Original {
			if obj.ExactIndexIn(hovering) == -1 {
				a.Selected = append(a.Selected, obj)
			}
		}
	} else {
		a.Selected = make([]*core.Selectable, len(hovering))
		copy(a.Selected, hovering)
	}
	a.SelectObjects(true, a.Selected...)
	a.ActionFinished(false, true, false)
}
func (a *Select) Act()    {}
func (a *Select) Cancel() {}
func (a *Select) Frame()  {}

func (a *Select) Undo() {
	a.SelectObjects(true, a.Original...)
}
func (a *Select) Redo() {
	a.SelectObjects(true, a.Selected...)
}
func (a *Select) RequiresLock() bool { return false }
