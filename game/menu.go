// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"io/fs"
	"log"
	"path/filepath"
	"strings"

	"github.com/gopxl/pixel/v2"

	"tlyakhov/gofoom/archetypes"
	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/controllers"
	"tlyakhov/gofoom/ecs"
	"tlyakhov/gofoom/ui"
)

var gameUI *ui.UI
var uiPageMain *ui.Page
var uiPageLoadWorld *ui.Page
var settingsNeedSave bool = false

func saveSettings() {
	ui.SaveSettings(constants.UserSettings, uiPageMain, uiPageSettings, uiPageKeyBindings)
}

func onWidgetChanged(w ui.IWidget) {
	settingsNeedSave = true
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
			&ui.Button{Widget: ui.Widget{Label: "Settings " + string(rune(16))}, Clicked: func(b *ui.Button) {
				gameUI.SetPage(uiPageSettings)
			}},
			&ui.Button{Widget: ui.Widget{Label: "Quit"},
				Clicked: func(b *ui.Button) {
					win.SetClosed(true)
				}},
		},
	}
	uiPageMain.Initialize()

	uiPageLoadWorld = &ui.Page{
		Parent:   uiPageMain,
		IsDialog: true,
		Title:    "Load World",
		Widgets: []ui.IWidget{
			&ui.Button{Widget: ui.Widget{Label: string(rune(17)) + " Main Menu"}, Clicked: func(b *ui.Button) {
				gameUI.SetPage(uiPageMain)
			}},
		}}

	filepath.Walk("data/worlds/", func(path string, info fs.FileInfo, err error) error {
		if !strings.HasSuffix(path, ".yaml") {
			return nil
		}
		name := filepath.Base(path)
		name = name[0 : len(name)-len(filepath.Ext(path))]
		b := &ui.Button{
			Widget: ui.Widget{
				Label:   name,
				Tooltip: path,
			},
			Clicked: func(b *ui.Button) {
				ecs.Initialize()
				if err := ecs.Load(path); err != nil {
					log.Printf("Error loading world %v: %v", path, err)
					return
				}
				archetypes.CreateFont("data/vga-font-8x8.png", "Default Font")
				renderer.Initialize()
				controllers.Respawn(true)
				ecs.Simulation.Integrate = integrateGame
				ecs.Simulation.Render = renderGame
				gameUI.Config.TextStyle = renderer.NewTextStyle()
				inMenu = false
				gameUI.SetPage(nil)
			}}
		uiPageLoadWorld.Widgets = append(uiPageLoadWorld.Widgets, b)
		return nil
	})
	uiPageLoadWorld.Initialize()

	initMenuOptions()
}

func menuInput() {
	if gameUI.Page == nil {
		gameUI.SetPage(uiPageMain)
	}

	if win.JustPressed(pixel.KeyW) || win.Repeated(pixel.KeyW) {
		gameUI.MoveUp()
	}
	if win.JustPressed(pixel.KeyS) || win.Repeated(pixel.KeyS) {
		gameUI.MoveDown()
	}
	if win.JustPressed(pixel.KeyA) || win.Repeated(pixel.KeyA) {
		gameUI.EditLeft()
	}
	if win.JustPressed(pixel.KeyD) || win.Repeated(pixel.KeyD) {
		gameUI.EditRight()
	}
	if win.JustReleased(pixel.KeyEnter) || win.JustReleased(pixel.KeySpace) {
		gameUI.Action()
	}
}
