// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"io/fs"
	"log"
	"path/filepath"
	"strings"

	"github.com/gopxl/pixel/v2"

	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/ui"
)

var gameUI *ui.UI
var uiPageMain *ui.Page
var uiPageLoadWorld *ui.Page

/*
	MaxPortals       = 100 // avoid infinite portal traversal
	MaxViewDistance         = 10000.0
	CollisionSteps          = 10
	MaxLightmapAge          = 3 // in frames
	LightmapRefreshDither   = 6 // in frames
*/

func saveSettings(w ui.IWidget) {
	ui.SaveSettings(constants.UserSettings, uiPageMain, uiPageOptions, uiPageKeyBindings)
}

func initializeMenus() {
	uiPageMain = &ui.Page{
		IsDialog: true,
		Title:    "Menu",
		Widgets: []ui.IWidget{
			&ui.Button{Widget: ui.Widget{Label: "Reset"}},
			&ui.Button{Widget: ui.Widget{Label: "Load World " + string(rune(16))}, Clicked: func(b *ui.Button) {
				gameUI.SetPage(uiPageLoadWorld)
			}},
			&ui.Button{Widget: ui.Widget{Label: "Options " + string(rune(16))}, Clicked: func(b *ui.Button) {
				gameUI.SetPage(uiPageOptions)
			}},
			&ui.Button{Widget: ui.Widget{Label: "Quit"},
				Clicked: func(b *ui.Button) {
					win.SetClosed(true)
				}},
		},
	}
	uiPageLoadWorld = &ui.Page{
		Parent:   uiPageMain,
		IsDialog: true,
		Title:    "Load World",
		Widgets: []ui.IWidget{
			&ui.Button{Widget: ui.Widget{Label: string(rune(17)) + " Main Menu"}, Clicked: func(b *ui.Button) {
				gameUI.SetPage(uiPageMain)
			}},
		}}

	filepath.Walk("data/", func(path string, info fs.FileInfo, err error) error {
		if !strings.HasSuffix(path, ".json") {
			return nil
		}
		b := &ui.Button{
			Widget: ui.Widget{
				Label:   "Load " + filepath.Base(path),
				Tooltip: path,
			},
			Clicked: func(b *ui.Button) {
				db.Clear()
				if err := db.Load(path); err != nil {
					log.Printf("Error loading world %v", err)
					return
				}
				db.Simulation.Integrate = integrateGame
				db.Simulation.Render = renderGame
				gameUI.Config.TextStyle = renderer.NewTextStyle()
				inMenu = false
				gameUI.SetPage(nil)
			}}
		uiPageLoadWorld.Widgets = append(uiPageLoadWorld.Widgets, b)
		return nil
	})

	initMenuOptions()
}

func menuInput() {
	if gameUI.Page == nil {
		gameUI.SetPage(uiPageMain)
	}

	if win.JustPressed(pixel.KeyW) || win.JustPressed(pixel.KeyUp) {
		gameUI.MoveUp()
	}
	if win.JustPressed(pixel.KeyS) || win.JustPressed(pixel.KeyDown) {
		gameUI.MoveDown()
	}
	if win.JustPressed(pixel.KeyA) || win.JustPressed(pixel.KeyLeft) {
		gameUI.EditLeft()
	}
	if win.JustPressed(pixel.KeyD) || win.JustPressed(pixel.KeyRight) {
		gameUI.EditRight()
	}
	if win.JustReleased(pixel.KeyEnter) || win.JustReleased(pixel.KeySpace) {
		gameUI.Action()
	}
}
