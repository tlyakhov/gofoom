package concepts

import (
	"reflect"
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

	if tLocal.Kind() == reflect.Ptr {
		tLocal = tLocal.Elem()
	}
	cm := ControllerMetadata{Type: tLocal, Methods: make(map[string]reflect.Value)}
	db.Controllers[tLocal.String()] = &cm
	for i := 0; i < tLocal.NumMethod(); i++ {
		m := tLocal.Method(i)
		cm.Methods[m.Name] = m.Func
	}
}

func (db *EntityComponentDB) NewControllerSet() *ControllerSet {
	result := &ControllerSet{
		EntityComponentDB: db,
		ByName:            make(map[string]Controller),
		All:               make([]Controller, len(DbTypes().Controllers)),
	}

	i := 0
	for _, t := range DbTypes().Controllers {
		c := reflect.New(t).Interface().(Controller)
		c.Parent(result)
		result.ByName[t.String()] = c
		result.All[i] = c
		i++
	}

	return result
}

func (set *ControllerSet) Act(target *EntityRef, source *EntityRef, method string) {
	for _, c := range set.All {
		if c.Target(target) && c.Source(source) {
			t := reflect.ValueOf(c).Type().String()
			DbTypes().Controllers[t].Methods[method].Call([]reflect.Value{reflect.ValueOf(c)})
		}
	}
}

func (set *ControllerSet) ActGlobal(method string) {
	for _, allComponents := range set.Components {
		for _, c := range allComponents {
			er := c.EntityRef()
			for _, controller := range set.All {
				if controller.Target(er) {
					t := reflect.ValueOf(controller).Type().String()
					DbTypes().Controllers[t].Methods[method].Call([]reflect.Value{reflect.ValueOf(controller)})
				}
			}
		}
	}
}
