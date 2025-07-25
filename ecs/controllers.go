// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"reflect"
	"sort"
)

// TODO: We originally had way more different types of controllers. Now there
// are only 3 methods. Can we make this simpler?
// TODO: Multiple controllers acting on the same ID should only iterate over the
// arena once.

func (types *typeMetadata) RegisterController(constructor func() Controller, priority int) {
	types.lock.Lock()
	defer types.lock.Unlock()
	instance := constructor()
	instanceType := reflect.ValueOf(instance).Type()
	types.Controllers = append(types.Controllers, controllerMetadata{
		Constructor: constructor,
		Type:        instanceType,
		Priority:    priority})
	sort.Slice(types.Controllers, func(i, j int) bool {
		return types.Controllers[i].Priority < types.Controllers[j].Priority
	})
}

// act calls a specific controller method on a component.
// It checks if the controller's target conditions are met for the component and
// its associated entities.
func act(controller Controller, component Attachable, method ControllerMethod) {
	if component.Base().Attachments == 1 {
		// If the component is attached to only one entity, check the target condition for that entity.
		if controller.Target(component, component.Base().Entity) {
			switch method {
			case ControllerAlways:
				controller.Always()
			case ControllerRecalculate:
				controller.Recalculate()
			}
		}
		return
	}
	// If the component is attached to multiple entities, iterate through them
	// and check the target condition for each.
	for _, e := range component.Base().Entities {
		if e == 0 {
			continue
		}
		if controller.Target(component, e) {
			switch method {
			case ControllerAlways:
				controller.Always()
			case ControllerRecalculate:
				controller.Recalculate()
			}
		}
	}
}

// Act runs controllers for a specific component, based on the provided component ID and method.
func Act(component Attachable, id ComponentID, method ControllerMethod) {
	for _, meta := range Types().Controllers {
		// Create a new instance of the controller.
		controller := meta.Constructor()
		// Check if the controller handles the specified method.
		if controller.Methods()&method == 0 {
			continue
		}
		// Check if the controller should be active based on the editor's pause state.
		if (Simulation.EditorPaused && controller.EditorPausedMethods()&method == 0) ||
			(!Simulation.EditorPaused && controller.Methods()&method == 0) {
			continue
		}
		// Check if the controller's component ID matches the provided ID.
		if controller.ComponentID() != id {
			continue
		}
		// Call the controller's method on the component.
		act(controller, component, method)
	}
}

// ActAllControllers runs all controllers for all components that have the specified method.
func ActAllControllers(method ControllerMethod) {
	for _, meta := range Types().Controllers {
		// Create a new instance of the controller.
		controller := meta.Constructor()
		// Check if the controller should be active based on the editor's pause state and if it handles the specified method.
		if (Simulation.EditorPaused && controller.EditorPausedMethods()&method == 0) ||
			(!Simulation.EditorPaused && controller.Methods()&method == 0) {
			continue
		}
		// Get the arena for the controller's component type.
		col := arenas[controller.ComponentID()]
		// Iterate through the components in the arena and call the controller's method on each active component.
		for i := range col.Cap() {
			if component := col.Attachable(i); component != nil {
				act(controller, component, method)
			}
		}
	}
}

// ActAllControllersOneEntity runs all controllers for a specific entity that have the specified method.
func ActAllControllersOneEntity(entity Entity, method ControllerMethod) {
	sid, local := localizeEntity(entity)
	if local == 0 {
		return
	}

	for _, meta := range Types().Controllers {
		// Create a new instance of the controller.
		controller := meta.Constructor()
		// Check if the controller should be active based on the editor's pause state and if it handles the specified method.
		if (Simulation.EditorPaused && controller.EditorPausedMethods()&method == 0) ||
			(!Simulation.EditorPaused && controller.Methods()&method == 0) {
			continue
		}
		// Iterate through the components attached to the entity.
		for _, component := range rows[sid][local] {
			if component == nil ||
				component.ComponentID() != controller.ComponentID() {
				continue
			}
			// Call the controller's method on the component.
			act(controller, component, method)
		}
	}
}
