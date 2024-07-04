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
	Original *core.Selection
	Selected *core.Selection
}

func (a *Select) OnMouseDown(evt *desktop.MouseEvent) {

	if evt.Modifier&fyne.KeyModifierShift != 0 {
		a.Modifier = SelectAdd
	} else if evt.Modifier&fyne.KeyModifierSuper != 0 {
		a.Modifier = SelectSub
	}

	a.Original = core.NewSelectionClone(a.State().SelectedObjects)

	a.Mode = "SelectionStart"
	a.SetMapCursor(desktop.TextCursor)
}

func (a *Select) OnMouseMove() {
	a.Mode = "Selecting"
}

func (a *Select) OnMouseUp() {
	hovering := a.State().HoveringObjects
	if hovering.Empty() { // User is trying to select a sector?
		hovering = core.NewSelection()
		for _, isector := range a.State().DB.AllOfType(core.SectorComponentIndex) {
			sector := isector.(*core.Sector)
			if sector.IsPointInside2D(&a.State().MouseWorld) {
				hovering.Add(core.SelectableFromSector(sector))
			}
		}
	}

	if a.Modifier == SelectAdd {
		a.Selected = core.NewSelectionClone(a.Original)
		for _, s := range hovering.Exact {
			a.Selected.Add(s)
		}
	} else if a.Modifier == SelectSub {
		a.Selected = core.NewSelection()
		for _, obj := range a.Original.Exact {
			if !hovering.Contains(obj) {
				a.Selected.Add(obj)
			}
		}
	} else {
		a.Selected = core.NewSelectionClone(hovering)
	}
	a.Selected.Normalize()
	a.SetSelection(true, a.Selected)
	a.ActionFinished(false, true, false)
}
func (a *Select) Act()    {}
func (a *Select) Cancel() {}
func (a *Select) Frame()  {}

func (a *Select) Undo() {
	a.SetSelection(true, a.Original)
}
func (a *Select) Redo() {
	a.SetSelection(true, a.Selected)
}
func (a *Select) RequiresLock() bool { return false }
