// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/controllers"
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
	toneMap := ecs.Singleton(materials.ToneMapCID).(*materials.ToneMap)
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
			toneMap.Precompute()
			// After everything's loaded, trigger the controllers
			ecs.ActAllControllers(ecs.ControllerPrecompute)
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
					Label:      "Controls " + string(rune(16)),
					LabelColor: &concepts.Vector4{1, 1, 0.3, 1},
					Tooltip:    "Keyboard input bindings, control options",
				},
				Clicked: func(b *ui.Button) {
					gameUI.SetPage(uiPageKeyBindings)
				}},
			&ui.Button{
				Widget: ui.Widget{
					Label:      "Apply & Save",
					LabelColor: &concepts.Vector4{1, 0.3, 0.3, 1},
					Tooltip:    "Apply and save settings",
				},
				Clicked: func(b *ui.Button) {
					uiPageSettings.Apply(uiPageSettings)
				}},
			&ui.Slider{
				Widget: ui.Widget{
					ID:      "gamma",
					Label:   "Brightness (Gamma)",
					Tooltip: "Default is 2.4",
					Justify: 1,
				},
				Min: 10, Max: 30, Value: int(toneMap.Gamma * 10), Step: 1,
			},
			&ui.Checkbox{
				Widget: ui.Widget{
					ID:      "multiRender",
					Label:   "Multithreaded Rendering",
					Tooltip: "Whether to parallelize rendering across multiple cores.\nHeavy impact on performance.",
					Justify: 1,
				},
				Value: renderer.Multithreaded,
			},
			&ui.Slider{
				Widget: ui.Widget{
					ID:      "multiBlocks",
					Label:   "Rendering threads",
					Tooltip: "Subdivide the screen into this many horizontal columns when rendering.\nIdeal seems to be ~2x physical cores. Heavy impact on performance.",
					Justify: 1,
				},
				Min: 2, Max: 64, Value: int(renderer.NumBlocks), Step: 2,
			},
			&ui.Slider{
				Widget: ui.Widget{
					ID:      "fov",
					Label:   "Field of View",
					Tooltip: "Angular extent of the player camera, in degrees.",
					Justify: 1,
				},
				Min: 45, Max: 160, Value: int(renderer.FOV), Step: 5,
			},
			&ui.Slider{
				Widget: ui.Widget{
					ID:      "lightGrid",
					Label:   "Lighting Fidelity",
					Tooltip: "Size of lightmap texels, in 1/10 of a world unit.\nLower is better, but impacts performance.",
					Justify: 1,
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
		Apply: func(p *ui.Page) {
			saveSettings()
		},
		Widgets: []ui.IWidget{
			&ui.Button{Widget: ui.Widget{Label: string(rune(17)) + " Options"}, Clicked: func(b *ui.Button) {
				gameUI.SetPage(uiPageSettings)
			}},
			&ui.Button{
				Widget: ui.Widget{
					Label:      "Apply & Save",
					LabelColor: &concepts.Vector4{1, 0.3, 0.3, 1},
					Tooltip:    "Apply and save settings",
				},
				Clicked: func(b *ui.Button) {
					uiPageKeyBindings.Apply(uiPageSettings)
				}},
			&ui.InputBinding{
				Widget: ui.Widget{
					ID:      "inputForward",
					Label:   "Move Forward",
					Tooltip: "Input for moving forward",
					Justify: -1,
				},
				EventID: controllers.EventIdForward,
				Input1:  "W",
				Input2:  "Up",
			},
			&ui.InputBinding{
				Widget: ui.Widget{
					ID:      "inputBack",
					Label:   "Move Backward",
					Tooltip: "Input for moving backward",
					Justify: -1,
				},
				EventID: controllers.EventIdBack,
				Input1:  "S",
				Input2:  "Down",
			},
			&ui.InputBinding{
				Widget: ui.Widget{
					ID:      "inputLeft",
					Label:   "Move Left",
					Tooltip: "Input for strafing left",
					Justify: -1,
				},
				EventID: controllers.EventIdLeft,
				Input1:  "A",
				Input2:  "Left",
			},
			&ui.InputBinding{
				Widget: ui.Widget{
					ID:      "inputRight",
					Label:   "Move Right",
					Tooltip: "Input for strafing right",
					Justify: -1,
				},
				EventID: controllers.EventIdRight,
				Input1:  "D",
				Input2:  "Right",
			},
			&ui.InputBinding{
				Widget: ui.Widget{
					ID:      "inputTurnLeft",
					Label:   "Turn Left",
					Tooltip: "Input for turning left",
					Justify: -1,
				},
				EventID: controllers.EventIdTurnLeft,
				Input1:  "Q",
				Input2:  "",
			},
			&ui.InputBinding{
				Widget: ui.Widget{
					ID:      "inputTurnRight",
					Label:   "Turn Right",
					Tooltip: "Input for turning right",
					Justify: -1,
				},
				EventID: controllers.EventIdTurnRight,
				Input1:  "E",
				Input2:  "",
			},
			&ui.InputBinding{
				Widget: ui.Widget{
					ID:      "inputUp",
					Label:   "Jump/Swim up",
					Tooltip: "Input for moving up",
					Justify: -1,
				},
				EventID: controllers.EventIdUp,
				Input1:  "Space",
				Input2:  "",
			},
			&ui.InputBinding{
				Widget: ui.Widget{
					ID:      "inputDown",
					Label:   "Crouch/Swim down",
					Tooltip: "Input for moving down",
					Justify: -1,
				},
				EventID: controllers.EventIdDown,
				Input1:  "C",
				Input2:  "",
			},
			&ui.InputBinding{
				Widget: ui.Widget{
					ID:      "inputPrimaryAction",
					Label:   "Primary/Fire",
					Tooltip: "Input for primary action, like firing weapon",
					Justify: -1,
				},
				EventID: controllers.EventIdPrimaryAction,
				Input1:  "MouseButton1",
				Input2:  "Enter",
			},
			&ui.InputBinding{
				Widget: ui.Widget{
					ID:      "inputSecondaryAction",
					Label:   "Secondary/Use",
					Tooltip: "Input for secondary action, like opening a door",
					Justify: -1,
				},
				EventID: controllers.EventIdSecondaryAction,
				Input1:  "MouseButton2",
				Input2:  "F",
			},
			&ui.InputBinding{
				Widget: ui.Widget{
					ID:      "inputYaw",
					Label:   "Yaw (turn axis)",
					Tooltip: "Mouse/joystick/gamepad axis for turning",
					Justify: -1,
				},
				EventID: controllers.EventIdYaw,
				Input1:  "MouseX",
				Input2:  "",
			},
			&ui.InputBinding{
				Widget: ui.Widget{
					ID:      "inputPitch",
					Label:   "Pitch (vertical axis)",
					Tooltip: "Mouse/joystick/gamepad axis for looking up/down",
					Justify: -1,
				},
				EventID: controllers.EventIdPitch,
				Input1:  "MouseY",
				Input2:  "",
			},
		},
	}
	uiPageKeyBindings.Initialize()
}
