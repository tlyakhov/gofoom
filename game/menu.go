// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"log"

	"github.com/gopxl/pixel/v2"

	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/ui"
)

var gameUI *ui.UI
var uiPageMain *ui.Page
var uiPageOptions *ui.Page
var uiPageKeyBindings *ui.Page

/*
	RenderBlocks     = 32  // When multi-threaded, each block will have its own goroutine
	MaxPortals       = 100 // avoid infinite portal traversal
	MaxViewDistance         = 10000.0
	FieldOfView             = 90
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
			&ui.Button{Widget: ui.Widget{Label: "Load World"}},
			&ui.Button{Widget: ui.Widget{Label: "Options " + string(rune(16))}, Clicked: func(b *ui.Button) {
				gameUI.SetPage(uiPageOptions)
			}},
			&ui.Button{Widget: ui.Widget{Label: "Quit"},
				Clicked: func(b *ui.Button) {
					win.SetClosed(true)
				}},
		},
	}
	uiPageOptions = &ui.Page{
		Parent:   uiPageMain,
		IsDialog: true,
		Title:    "Options",
		Widgets: []ui.IWidget{
			&ui.Button{Widget: ui.Widget{Label: string(rune(17)) + " Main Menu"}, Clicked: func(b *ui.Button) {
				gameUI.SetPage(uiPageMain)
			}},
			&ui.Button{
				Widget: ui.Widget{
					Label:   "Controls " + string(rune(16)),
					Tooltip: "Keyboard input bindings, control options",
				},
				Clicked: func(b *ui.Button) {
					gameUI.SetPage(uiPageKeyBindings)
				}},
			&ui.Checkbox{
				Widget: ui.Widget{
					ID:      "multiRender",
					Label:   "Multithreaded Rendering",
					Tooltip: "Whether to parallelize rendering across multiple cores.\nHeavy impact on performance.",
				},
				Value: renderer.Multithreaded,
				Checked: func(cb *ui.Checkbox) {
					renderer.Multithreaded = cb.Value
				}},
			&ui.Slider{
				Widget: ui.Widget{
					ID:      "multiBlocks",
					Label:   "Rendering threads",
					Tooltip: "Subdivide the screen into this many horizontal columns when rendering.\nIdeal seems to be ~2x physical cores. Heavy impact on performance.",
				},
				Min: 2, Max: 64, Value: int(renderer.Blocks), Step: 2,
				Moved: func(s *ui.Slider) {
					renderer.Blocks = s.Value
					renderer.Initialize()
				},
			},
			&ui.Slider{
				Widget: ui.Widget{
					ID:      "fov",
					Label:   "Field of View",
					Tooltip: "Angular extent of the player camera, in degrees.",
				},
				Min: 45, Max: 160, Value: int(renderer.FOV), Step: 5,
				Moved: func(s *ui.Slider) {
					//	renderer.RenderLock.Lock()
					//	defer renderer.RenderLock.Unlock()
					renderer.FOV = float64(s.Value)
					renderer.Initialize()
				},
			},
			&ui.Slider{
				Widget: ui.Widget{
					ID:      "lightGrid",
					Label:   "Lighting Fidelity",
					Tooltip: "Size of lightmap texels, in 1/10 of a world unit.\nLower is better, but impacts performance.",
				},
				Min: 5, Max: 100, Value: int(renderer.LightGrid * 10), Step: 1,
				Moved: func(s *ui.Slider) {
					//	renderer.RenderLock.Lock()
					//	defer renderer.RenderLock.Unlock()
					renderer.LightGrid = float64(s.Value) / 10.0
					log.Printf("%v", renderer.LightGrid)
					renderer.Initialize()
					// After everything's loaded, trigger the controllers
					db.ActAllControllers(concepts.ControllerRecalculate)
				},
			},
		},
	}
	uiPageKeyBindings = &ui.Page{
		Parent:   uiPageOptions,
		IsDialog: true,
		Title:    "Key Bindings",
		Widgets: []ui.IWidget{
			&ui.Button{Widget: ui.Widget{Label: string(rune(17)) + " Options"}, Clicked: func(b *ui.Button) {
				gameUI.SetPage(uiPageOptions)
			}},
		},
	}
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
