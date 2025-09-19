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

func processBinding(name string, eventID dynamic.EventID) {
	binding := uiPageKeyBindings.Mapped[name].(*ui.InputBinding)
	if binding.Input1 != "" {
		processBindingInput(eventID, binding.Input1)
	}
	if binding.Input2 != "" {
		processBindingInput(eventID, binding.Input2)
	}
}

// TODO: unify this with editor.
// TODO: Improve Mouse look (or any axis-based bindings) by relying on deltas
// rather than absolute amounts.
func gameInput() {
	// TODO: This is a lot of polling. Instead, let's use the callback approach
	// the pixel package provides, e.g. win.SetButtonCallback
	processBinding("inputForward", controllers.EventIdForward)
	processBinding("inputBack", controllers.EventIdBack)
	processBinding("inputLeft", controllers.EventIdLeft)
	processBinding("inputRight", controllers.EventIdRight)
	processBinding("inputTurnLeft", controllers.EventIdTurnLeft)
	processBinding("inputTurnRight", controllers.EventIdTurnRight)
	processBinding("inputUp", controllers.EventIdUp)
	processBinding("inputDown", controllers.EventIdDown)
	processBinding("inputPrimaryAction", controllers.EventIdPrimaryAction)
	processBinding("inputSecondaryAction", controllers.EventIdSecondaryAction)
	processBinding("inputYaw", controllers.EventIdYaw)
	processBinding("inputPitch", controllers.EventIdPitch)
}
