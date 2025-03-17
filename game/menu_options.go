// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/ecs"
	"tlyakhov/gofoom/ui"
)

/*
TODO: Also add these:

MaxPortals       = 100 // avoid infinite portal traversal
MaxViewDistance         = 10000.0
CollisionSteps          = 10
MaxLightmapAge          = 3 // in frames
LightmapRefreshDither   = 6 // in frames
*/
var uiPageSettings, uiPageKeyBindings *ui.Page

func initMenuOptions() {
	toneMap := u.Singleton(materials.ToneMapCID).(*materials.ToneMap)
	uiPageSettings = &ui.Page{
		Parent:   uiPageMain,
		IsDialog: true,
		Title:    "Options",
		Apply: func(p *ui.Page) {
			renderer.Multithreaded = p.Widget("multiRender").(*ui.Checkbox).Value
			renderer.NumBlocks = p.Widget("multiBlocks").(*ui.Slider).Value
			renderer.FOV = float64(p.Widget("fov").(*ui.Slider).Value)
			renderer.LightGrid = float64(p.Widget("lightGrid").(*ui.Slider).Value) / 10.0
			toneMap.Gamma = float64(p.Widget("gamma").(*ui.Slider).Value) / 10.0
			toneMap.Recalculate()
			// After everything's loaded, trigger the controllers
			u.ActAllControllers(ecs.ControllerRecalculate)
			renderer.Initialize()
			saveSettings()
		},
		Widgets: []ui.IWidget{
			&ui.Button{Widget: ui.Widget{Label: string(rune(17)) + " Main Menu"}, Clicked: func(b *ui.Button) {
				if settingsNeedSave {
					// TODO: Ask user to save or discard
				}
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
			&ui.Button{
				Widget: ui.Widget{
					Label:   "Apply & Save",
					Tooltip: "Apply and save settings",
				},
				Clicked: func(b *ui.Button) {
					uiPageSettings.Apply(uiPageSettings)
				}},
			&ui.Slider{
				Widget: ui.Widget{
					ID:      "gamma",
					Label:   "Brightness (Gamma)",
					Tooltip: "Default is 2.4",
				},
				Min: 10, Max: 30, Value: int(toneMap.Gamma * 10), Step: 1,
			},
			&ui.Checkbox{
				Widget: ui.Widget{
					ID:      "multiRender",
					Label:   "Multithreaded Rendering",
					Tooltip: "Whether to parallelize rendering across multiple cores.\nHeavy impact on performance.",
				},
				Value: renderer.Multithreaded,
			},
			&ui.Slider{
				Widget: ui.Widget{
					ID:      "multiBlocks",
					Label:   "Rendering threads",
					Tooltip: "Subdivide the screen into this many horizontal columns when rendering.\nIdeal seems to be ~2x physical cores. Heavy impact on performance.",
				},
				Min: 2, Max: 64, Value: int(renderer.NumBlocks), Step: 2,
			},
			&ui.Slider{
				Widget: ui.Widget{
					ID:      "fov",
					Label:   "Field of View",
					Tooltip: "Angular extent of the player camera, in degrees.",
				},
				Min: 45, Max: 160, Value: int(renderer.FOV), Step: 5,
			},
			&ui.Slider{
				Widget: ui.Widget{
					ID:      "lightGrid",
					Label:   "Lighting Fidelity",
					Tooltip: "Size of lightmap texels, in 1/10 of a world unit.\nLower is better, but impacts performance.",
				},
				Min: 5, Max: 100, Value: int(renderer.LightGrid * 10), Step: 1,
			},
		},
	}
	uiPageSettings.Initialize()

	uiPageKeyBindings = &ui.Page{
		Parent:   uiPageSettings,
		IsDialog: true,
		Title:    "Key Bindings",
		Widgets: []ui.IWidget{
			&ui.Button{Widget: ui.Widget{Label: string(rune(17)) + " Options"}, Clicked: func(b *ui.Button) {
				gameUI.SetPage(uiPageSettings)
			}},
		},
	}
	uiPageKeyBindings.Initialize()
}
