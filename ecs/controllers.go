// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"reflect"
	"sort"
)

// TODO: We originally had way more different types of controllers. Now there
// are only 3 methods. Can we make this simpler?

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

func act(controller Controller, component Attachable, method ControllerMethod) {
	if component.Base().Attachments == 1 {
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

func (db *ECS) Act(component Attachable, id ComponentID, method ControllerMethod) {
	for _, meta := range Types().Controllers {
		controller := meta.Constructor()
		if controller.Methods()&method == 0 {
			continue
		}
		if (db.EditorPaused && controller.EditorPausedMethods()&method == 0) ||
			(!db.EditorPaused && controller.Methods()&method == 0) {
			continue
		}
		if controller.ComponentID() != id {
			continue
		}
		act(controller, component, method)
	}
}

func (db *ECS) ActAllControllers(method ControllerMethod) {
	for _, meta := range Types().Controllers {
		controller := meta.Constructor()
		if (db.EditorPaused && controller.EditorPausedMethods()&method == 0) ||
			(!db.EditorPaused && controller.Methods()&method == 0) {
			continue
		}
		col := db.columns[controller.ComponentID()]
		for i := range col.Cap() {
			if component := col.Attachable(i); component != nil {
				act(controller, component, method)
			}
		}
	}
}

func (db *ECS) ActAllControllersOneEntity(entity Entity, method ControllerMethod) {
	if entity == 0 || len(db.rows) <= int(entity) {
		return
	}

	for _, meta := range Types().Controllers {
		controller := meta.Constructor()
		if (db.EditorPaused && controller.EditorPausedMethods()&method == 0) ||
			(!db.EditorPaused && controller.Methods()&method == 0) {
			continue
		}
		for _, component := range db.rows[entity] {
			if component == nil ||
				component.Base().ComponentID != controller.ComponentID() {
				continue
			}
			act(controller, component, method)
		}
	}
}
