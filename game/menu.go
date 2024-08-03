// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"github.com/gopxl/pixel/v2"

	"tlyakhov/gofoom/ui"
)

var gameUI *ui.UI
var uiPageMain *ui.Page

func InitializeMenus() {
	uiPageMain = &ui.Page{
		Items: []ui.IWidget{
			&ui.Button{Widget: ui.Widget{Label: "Reset"}},
			&ui.Button{Widget: ui.Widget{Label: "Load World"}},
			&ui.Button{Widget: ui.Widget{Label: "Options"}},
			&ui.Button{Widget: ui.Widget{Label: "Quit", Action: func() {
				win.SetClosed(true)
			}}},
		},
	}
	gameUI.SetPage(uiPageMain)
}

func menuInput() {
	if gameUI.Page == nil {
		gameUI.SetPage(uiPageMain)
	}

	if win.JustPressed(pixel.KeyW) || win.JustPressed(pixel.KeyUp) {
		gameUI.Page.SelectedItem--
		if gameUI.Page.SelectedItem < 0 {
			gameUI.Page.SelectedItem = 0
		}
	}
	if win.JustPressed(pixel.KeyS) || win.JustPressed(pixel.KeyDown) {
		gameUI.Page.SelectedItem++
		if gameUI.Page.SelectedItem >= len(gameUI.Page.Items) {
			gameUI.Page.SelectedItem = len(gameUI.Page.Items) - 1
		}
	}
	if win.JustReleased(pixel.KeyEnter) || win.JustReleased(pixel.KeySpace) {
		item := gameUI.Page.Items[gameUI.Page.SelectedItem]
		if item.GetWidget().Action != nil {
			item.GetWidget().Action()
		}
	}
}
