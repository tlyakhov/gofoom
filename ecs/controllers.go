// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"reflect"
	"sort"
)

// TODO: We originally had way more different types of controllers. Now there
// are only 2 methods. Can we make this simpler?
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
func act(controller Controller, component Component, f func()) {
	if component.Base().Attachments == 1 {
		// If the component is attached to only one entity, check the target condition for that entity.
		if controller.Target(component, component.Base().Entity) {
			f()
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
			f()
		}
	}

	// if the component is attached with indirects, iterate through them
	// and check the target condition for each.
	cwi, ok := component.(ComponentWithIndirects)
	if !ok {
		return
	}
	for _, e := range *cwi.Indirects() {
		if e == 0 {
			continue
		}
		if controller.Target(component, e) {
			f()
		}
	}
}

// Act runs controllers for a specific component, based on the provided component ID and method.
func Act(component Component, id ComponentID, method ControllerMethod) {
	var f func()

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
		switch method {
		case ControllerFrame:
			f = controller.Frame
		case ControllerPrecompute:
			f = controller.Precompute
		}
		// Call the controller's method on the component.
		act(controller, component, f)
	}
}

// ActAllControllers runs all controllers for all components that have the specified method.
func ActAllControllers(method ControllerMethod) {
	var f func()
	for _, meta := range Types().Controllers {
		// Create a new instance of the controller.
		controller := meta.Constructor()
		// Check if the controller should be active based on the editor's pause state and if it handles the specified method.
		if (Simulation.EditorPaused && controller.EditorPausedMethods()&method == 0) ||
			(!Simulation.EditorPaused && controller.Methods()&method == 0) {
			continue
		}
		switch method {
		case ControllerFrame:
			f = controller.Frame
		case ControllerPrecompute:
			f = controller.Precompute
		}
		// Get the arena for the controller's component type.
		arena := arenas[controller.ComponentID()]
		// Iterate through the components in the arena and call the controller's method on each active component.
		for i := range arena.Cap() {
			if component := arena.Component(i); component != nil {
				act(controller, component, f)
			}
		}
	}
}

// ActAllControllersOneEntity runs all controllers for a specific entity that have the specified method.
func ActAllControllersOneEntity(entity Entity, method ControllerMethod) {
	sid, local := localizeEntityAndCheckRange(entity)
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
		var f func()
		switch method {
		case ControllerFrame:
			f = controller.Frame
		case ControllerPrecompute:
			f = controller.Precompute
		}
		// Iterate through the components attached to the entity.
		for _, component := range rows[sid][local] {
			if component == nil ||
				component.ComponentID() != controller.ComponentID() {
				continue
			}
			// Call the controller's method on the component.
			act(controller, component, f)
		}
	}
}
