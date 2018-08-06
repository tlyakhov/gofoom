package registry

import (
	"reflect"
	"sync"
)

type typeRegistry struct {
	All map[string]reflect.Type
}

var instance *typeRegistry
var once sync.Once

func Instance() *typeRegistry {
	once.Do(func() {
		instance = &typeRegistry{
			All: make(map[string]reflect.Type),
		}
	})
	return instance
}

func (tr *typeRegistry) Register(local interface{}) {
	tLocal := reflect.ValueOf(local).Type()

	if tLocal.Kind() == reflect.Ptr {
		tLocal = tLocal.Elem()
	}

	tr.All[tLocal.String()] = tLocal
	tr.All[reflect.PtrTo(tLocal).String()] = reflect.PtrTo(tLocal)
}

func Type(name string) reflect.Type {
	return Instance().All[name]
}
