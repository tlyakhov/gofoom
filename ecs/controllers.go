// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"reflect"
	"sort"
)

// TODO: A lot of unnecessary complexity here because we originally had way more
// different types of controllers. Now there are only really 2 methods: "Loaded"
// and "Recalculate". Should simplify.

func (types *typeMetadata) RegisterController(instance Controller, priority int) {
	types.lock.Lock()
	defer types.lock.Unlock()
	instanceType := reflect.ValueOf(instance).Type()

	if instanceType.Kind() != reflect.Ptr {
		panic("Attempt to register controller with a value type (should be a pointer)")
	}
	types.Controllers = append(types.Controllers, controllerMetadata{
		Controller: instance,
		Type:       instanceType,
		Priority:   priority})
	sort.Slice(types.Controllers, func(i, j int) bool {
		return types.Controllers[i].Priority < types.Controllers[j].Priority
	})
}

func act(controller Controller, method ControllerMethod) {
	switch method {
	case ControllerAlways:
		controller.Always()
	case ControllerLoaded:
		controller.Loaded()
	case ControllerRecalculate:
		controller.Recalculate()
	}
}

func (db *ECS) Act(component Attachable, index int, method ControllerMethod) {
	for _, meta := range Types().Controllers {
		controller := reflect.New(meta.Type.Elem()).Interface().(Controller)
		if controller == nil || controller.Methods()&method == 0 {
			continue
		}
		if controller.Methods()&method == 0 ||
			controller.ComponentIndex() != index ||
			!controller.Target(component) {
			continue
		}
		act(controller, method)
	}
}

func (db *ECS) ActAllControllers(method ControllerMethod) {
	for _, meta := range Types().Controllers {
		controller := reflect.New(meta.Type.Elem()).Interface().(Controller)
		if controller == nil || controller.Methods()&method == 0 {
			continue
		}

		for _, c := range db.components[controller.ComponentIndex()] {
			if !controller.Target(c) {
				continue
			}
			act(controller, method)
		}
	}
}
