// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"log"
	"tlyakhov/gofoom/editor/state"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

type Placeable interface {
	BeginPoint(m fyne.KeyModifier, button desktop.MouseButton) bool
	Point() bool
	EndPoint() bool
	Placing() bool
}

type Place struct {
	state.Action

	Mode     string
	Modifier fyne.KeyModifier
}

func (a *Place) BeginPoint(m fyne.KeyModifier, button desktop.MouseButton) bool {
	log.Printf("Place.BeginPoint: Mode = %v", a.Mode)
	if a.Mode != "" {
		return false
	}

	a.Modifier = m
	a.Mode = "Placing"
	a.SetMapCursor(desktop.CrosshairCursor)
	return true
}

func (a *Place) Point() bool {
	log.Printf("Place.Point: Mode = %v", a.Mode)
	return a.Mode != ""

}

func (a *Place) EndPoint() bool {
	log.Printf("Place.EndPoint: Mode = %v", a.Mode)
	if a.Mode == "" {
		return false
	}
	a.Mode = ""
	return true
}

func (a *Place) Placing() bool {
	return a.Mode != ""
}
