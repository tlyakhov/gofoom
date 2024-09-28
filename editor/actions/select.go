// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/selection"
	"tlyakhov/gofoom/ecs"

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
	Place

	Modifier SelectModifier
	Original *selection.Selection
	Selected *selection.Selection
}

func (a *Select) BeginPoint(m fyne.KeyModifier, b desktop.MouseButton) bool {
	if !a.Place.BeginPoint(m, b) {
		return false
	}

	if m&fyne.KeyModifierShift != 0 {
		a.Modifier = SelectAdd
	} else if m&fyne.KeyModifierSuper != 0 {
		a.Modifier = SelectSub
	}

	a.Original = selection.NewSelectionClone(a.State().SelectedObjects)
	return true
}

func (a *Select) EndPoint() bool {
	if !a.Place.EndPoint() {
		return false
	}

	hovering := a.State().HoveringObjects
	if hovering.Empty() { // User is trying to select a sector?
		hovering = selection.NewSelection()
		col := ecs.ColumnFor[core.Sector](a.State().ECS, core.SectorCID)
		for i := range col.Cap() {
			sector := col.Value(i)
			if sector == nil {
				continue
			}
			if sector.IsPointInside2D(&a.State().MouseWorld) {
				hovering.Add(selection.SelectableFromSector(sector))
			}
		}
	}

	if a.Modifier == SelectAdd {
		a.Selected = selection.NewSelectionClone(a.Original)
		for _, s := range hovering.Exact {
			a.Selected.Add(s)
		}
	} else if a.Modifier == SelectSub {
		a.Selected = selection.NewSelection()
		for _, obj := range a.Original.Exact {
			if !hovering.Contains(obj) {
				a.Selected.Add(obj)
			}
		}
	} else {
		a.Selected = selection.NewSelectionClone(hovering)
	}
	a.Selected.Normalize()
	a.SetSelection(true, a.Selected)
	a.ActionFinished(false, true, false)
	return true
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
