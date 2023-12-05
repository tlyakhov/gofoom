package concepts

import (
	"reflect"
	"sort"
)

type ControllerSet struct {
	*EntityComponentDB
	ByName map[string]Controller
	All    []Controller
}

func (db *dbTypes) RegisterController(local any) {
	db.lock.Lock()
	defer db.lock.Unlock()
	tLocal := reflect.ValueOf(local).Type()

	if tLocal.Kind() != reflect.Ptr {
		panic("Attempt to register controller with a value type (should be a pointer)")
	}
	db.Controllers[tLocal.String()] = tLocal
}

func (db *EntityComponentDB) NewControllerSet() *ControllerSet {
	result := &ControllerSet{
		EntityComponentDB: db,
		ByName:            make(map[string]Controller),
		All:               make([]Controller, len(DbTypes().Controllers)),
	}

	i := 0
	for _, t := range DbTypes().Controllers {
		c := reflect.New(t.Elem()).Interface().(Controller)
		c.Parent(result)
		result.ByName[t.String()] = c
		result.All[i] = c
		i++
	}
	sort.Slice(result.All, func(i, j int) bool {
		return result.All[i].Priority() < result.All[j].Priority()
	})
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

func (set *ControllerSet) Act(target *EntityRef, method ControllerMethod) {
	for _, c := range set.All {
		if c.Methods()&method == 0 {
			continue
		}
		if c.Target(target) {
			act(c, method)
		}
	}
}

func (set *ControllerSet) ActGlobal(method ControllerMethod) {
	for _, controller := range set.All {
		if controller.Methods()&method == 0 {
			continue
		}
		for _, allComponents := range set.Components {
			for _, component := range allComponents {
				er := component.Ref()
				if controller.Target(er) {
					act(controller, method)
				}
			}
		}
	}
}
