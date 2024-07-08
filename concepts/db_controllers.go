// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package concepts

import (
	"reflect"
	"sort"
)

type ControllerSet struct {
	*EntityComponentDB
	All []Controller
}

func (dbt *dbTypes) RegisterController(local Controller) {
	dbt.lock.Lock()
	defer dbt.lock.Unlock()
	tLocal := reflect.ValueOf(local).Type()

	if tLocal.Kind() != reflect.Ptr {
		panic("Attempt to register controller with a value type (should be a pointer)")
	}
	dbt.Controllers = append(dbt.Controllers, tLocal)
	// This is incredibly inefficient, but our N count is so low it doesn't
	// matter. Can refactor later if necessary
	sort.Slice(dbt.Controllers, func(i, j int) bool {
		ic := reflect.New(dbt.Controllers[i].Elem()).Interface().(Controller)
		jc := reflect.New(dbt.Controllers[j].Elem()).Interface().(Controller)
		return ic.Priority() < jc.Priority()
	})
}

func (db *EntityComponentDB) NewControllerSet() *ControllerSet {
	result := &ControllerSet{
		EntityComponentDB: db,
		All:               make([]Controller, len(DbTypes().Controllers)),
	}

	for i, t := range DbTypes().Controllers {
		c := reflect.New(t.Elem()).Interface().(Controller)
		c.Parent(result)
		result.All[i] = c
	}
	return result
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

func (set *ControllerSet) Act(component Attachable, index int, method ControllerMethod) {
	for _, controller := range set.All {
		if controller.Methods()&method == 0 ||
			controller.ComponentIndex() != index ||
			!controller.Target(component) {
			continue
		}
		act(controller, method)
	}
}

func (db *EntityComponentDB) ActAllControllers(method ControllerMethod) {
	set := db.NewControllerSet()
	for _, controller := range set.All {
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
