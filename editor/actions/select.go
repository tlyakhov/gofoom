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

// Declare conformity with interfaces
var _ fyne.Draggable = (*Select)(nil)
var _ fyne.SecondaryTappable = (*Select)(nil)
var _ desktop.Mouseable = (*Select)(nil)

type Select struct {
	state.IEditor

	Mode     string
	Modifier SelectModifier
	Original *core.Selection
	Selected *core.Selection
}

// TODO: This begin/end pattern for action could be generalized to avoid
// so much similar code across the editor actions.
func (a *Select) begin(m fyne.KeyModifier) {
	if a.Mode != "" {
		return
	}

	if m&fyne.KeyModifierShift != 0 {
		a.Modifier = SelectAdd
	} else if m&fyne.KeyModifierSuper != 0 {
		a.Modifier = SelectSub
	}

	a.Original = core.NewSelectionClone(a.State().SelectedObjects)

	a.Mode = "Selecting"
	a.SetMapCursor(desktop.TextCursor)
}

func (a *Select) end() {
	if a.Mode != "Selecting" {
		return
	}

	hovering := a.State().HoveringObjects
	if hovering.Empty() { // User is trying to select a sector?
		hovering = core.NewSelection()
		for _, isector := range a.State().ECS.AllOfType(core.SectorComponentIndex) {
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
	a.Mode = ""
	a.SetSelection(true, a.Selected)
	a.ActionFinished(false, true, false)
}

func (a *Select) MouseDown(evt *desktop.MouseEvent) {
	a.begin(evt.Modifier)
}

func (a *Select) TappedSecondary(evt *fyne.PointEvent) {
	if a.Mode == "" {
		a.begin(fyne.KeyModifier(0))
		a.end()
	}
}

func (a *Select) Dragged(evt *fyne.DragEvent) {
	a.begin(fyne.KeyModifier(0))
}

func (a *Select) DragEnd() {
	a.end()
}

func (a *Select) MouseUp(evt *desktop.MouseEvent) {
	a.end()
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

func (a *Select) Status() string {
	return "Selecting"
}
