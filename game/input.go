// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"tlyakhov/gofoom/controllers"
	"tlyakhov/gofoom/dynamic"
	"tlyakhov/gofoom/ecs"
	"tlyakhov/gofoom/ui"

	"github.com/gopxl/pixel/v2"
)

func processBindingInput(eventID dynamic.EventID, input string) {
	if b, ok := buttonBindings[input]; ok {
		if win.Pressed(b) {
			ecs.Simulation.NewEvent(eventID, &controllers.EntityEventParams{Entity: renderer.Player.Entity})
		}
		return
	}
	if b, ok := gamepadButtonBindings[input]; ok {
		if win.JoystickPressed(pixel.Joystick1, b) {
			ecs.Simulation.NewEvent(eventID, &controllers.EntityEventParams{Entity: renderer.Player.Entity})
		}
		return
	}
	if b, ok := gamepadAxisBindings[input]; ok {
		axis := win.JoystickAxis(pixel.Joystick1, b)
		ecs.Simulation.NewEvent(eventID, &controllers.EntityAxisEventParams{
			Entity:    renderer.Player.Entity,
			AxisValue: axis,
		})
		return
	}
	if _, ok := mouseAxisBindings[input]; ok {
		var axis float64
		switch input {
		case "MouseX":
			axis = win.MousePosition().X - win.MousePreviousPosition().X
		case "MouseY":
			axis = win.MousePosition().Y - win.MousePreviousPosition().Y
		case "MouseScrollX":
			axis = win.MouseScroll().X - win.MousePreviousScroll().X
		case "MouseScrollY":
			axis = win.MouseScroll().Y - win.MousePreviousScroll().Y
		}
		ecs.Simulation.NewEvent(eventID, &controllers.EntityAxisEventParams{
			Entity:    renderer.Player.Entity,
			AxisValue: axis,
		})
		return
	}
}

func gameInput() {
	for _, w := range uiPageKeyBindings.Widgets {
		if binding, ok := w.(*ui.InputBinding); ok {
			if binding.Input1 != "" {
				processBindingInput(binding.EventID, binding.Input1)
			}
			if binding.Input2 != "" {
				processBindingInput(binding.EventID, binding.Input2)
			}
		}
	}
}
