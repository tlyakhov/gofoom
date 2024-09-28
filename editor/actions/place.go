// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"tlyakhov/gofoom/editor/state"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

type Placeable interface {
	BeginPoint(m fyne.KeyModifier, button desktop.MouseButton) bool
	Point() bool
	EndPoint() bool
}

type Place struct {
	state.IEditor

	Mode     string
	Modifier fyne.KeyModifier
}

func (a *Place) BeginPoint(m fyne.KeyModifier, button desktop.MouseButton) bool {
	if a.Mode != "" {
		return false
	}

	a.Modifier = m
	a.Mode = "Begin"
	a.SetMapCursor(desktop.TextCursor)
	return true
}

func (a *Place) Point() bool {
	return a.Mode != ""

}

func (a *Place) EndPoint() bool {
	if a.Mode == "" {
		return false
	}
	a.Mode = ""
	return true
}
