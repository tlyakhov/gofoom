// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"tlyakhov/gofoom/render"

	"github.com/gopxl/pixel/v2"
)

var ui *render.UI
var menuMain *render.UIPage

func InitializeMenus() {
	menuMain = &render.UIPage{
		Items: []render.UIElement{
			{Label: "Reset"},
			{Label: "Load World"},
			{Label: "Options"},
			{Label: "Quit", Action: func() {
				win.SetClosed(true)
			}},
		},
	}
	ui.Page = menuMain
}

func menuInput() {
	if ui.Page == nil {
		ui.Page = menuMain
	}

	if win.JustReleased(pixel.KeyW) || win.JustReleased(pixel.KeyUp) {
		ui.Page.SelectedItem--
		if ui.Page.SelectedItem < 0 {
			ui.Page.SelectedItem = 0
		}
	}
	if win.JustReleased(pixel.KeyS) || win.JustReleased(pixel.KeyDown) {
		ui.Page.SelectedItem++
		if ui.Page.SelectedItem >= len(ui.Page.Items) {
			ui.Page.SelectedItem = len(ui.Page.Items) - 1
		}
	}
	if win.JustReleased(pixel.KeyEnter) || win.JustReleased(pixel.KeySpace) {
		item := ui.Page.Items[ui.Page.SelectedItem]
		if item.Action != nil {
			item.Action()
		}
	}
}
