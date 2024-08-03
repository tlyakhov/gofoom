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
		Widgets: []ui.IWidget{
			&ui.Button{Widget: ui.Widget{Label: "Reset"}},
			&ui.Button{Widget: ui.Widget{Label: "Load World"}},
			&ui.Button{Widget: ui.Widget{Label: "Options"}},
			&ui.Checkbox{Widget: ui.Widget{Label: "Check me"},
				Checked: func(cb *ui.Checkbox) {
				}},
			&ui.Slider{Widget: ui.Widget{Label: "Slide me"},
				Min: 0, Max: 100, Value: 50, Step: 5,
				Moved: func(s *ui.Slider) {
				}},
			&ui.Button{Widget: ui.Widget{Label: "Quit"},
				Clicked: func(b *ui.Button) {
					win.SetClosed(true)
				}},
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
		if gameUI.Page.SelectedItem >= len(gameUI.Page.Widgets) {
			gameUI.Page.SelectedItem = len(gameUI.Page.Widgets) - 1
		}
	}
	if win.JustPressed(pixel.KeyA) || win.JustPressed(pixel.KeyLeft) {
		switch w := gameUI.Page.Widgets[gameUI.Page.SelectedItem].(type) {
		case *ui.Slider:
			w.Value -= w.Step
			if w.Value < w.Min {
				w.Value = w.Min
			}
			if w.Moved != nil {
				w.Moved(w)
			}
		}
	}
	if win.JustPressed(pixel.KeyD) || win.JustPressed(pixel.KeyRight) {
		switch w := gameUI.Page.Widgets[gameUI.Page.SelectedItem].(type) {
		case *ui.Slider:
			w.Value += w.Step
			if w.Value > w.Max {
				w.Value = w.Max
			}
			if w.Moved != nil {
				w.Moved(w)
			}
		}
	}
	if win.JustReleased(pixel.KeyEnter) || win.JustReleased(pixel.KeySpace) {
		switch w := gameUI.Page.Widgets[gameUI.Page.SelectedItem].(type) {
		case *ui.Button:
			if w.Clicked != nil {
				w.Clicked(w)
			}
		case *ui.Checkbox:
			w.IsChecked = !w.IsChecked
			if w.Checked != nil {
				w.Checked(w)
			}
		}
	}
}
