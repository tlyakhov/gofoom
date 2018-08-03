package registry

import (
	"fmt"
	"reflect"
	"sync"
)

type typeRegistry struct {
	ByPackage map[string]map[string]reflect.Type
	All       map[string]reflect.Type
}

var instance *typeRegistry
var once sync.Once

func Instance() *typeRegistry {
	once.Do(func() {
		instance = &typeRegistry{
			ByPackage: make(map[string]map[string]reflect.Type),
			All:       make(map[string]reflect.Type),
		}
	})
	return instance
}

func (tr *typeRegistry) Register(local interface{}) {
	tr.RegisterMapped(local, local)
}

func (tr *typeRegistry) RegisterMapped(local, external interface{}) {
	tLocal := reflect.ValueOf(local).Type()
	tExternal := reflect.ValueOf(external).Type()

	if tLocal.Kind() == reflect.Ptr {
		tLocal = tExternal.Elem()
	}
	if tExternal.Kind() == reflect.Ptr {
		tExternal = tExternal.Elem()
	}

	tr.All[tLocal.String()] = tLocal
	tr.All[reflect.PtrTo(tLocal).String()] = reflect.PtrTo(tLocal)

	pkgLocal := tLocal.PkgPath()
	pkgExternal := tExternal.PkgPath()

	var tmPkgLocal map[string]reflect.Type
	var tmPkgExternal map[string]reflect.Type
	var ok bool

	if tmPkgLocal, ok = tr.ByPackage[pkgLocal]; !ok {
		tmPkgLocal = make(map[string]reflect.Type)
		tr.ByPackage[pkgLocal] = tmPkgLocal
	}

	if pkgLocal != pkgExternal {
		if tmPkgExternal, ok = tr.ByPackage[pkgExternal]; !ok {
			tmPkgExternal = make(map[string]reflect.Type)
			tr.ByPackage[pkgExternal] = tmPkgExternal
		}
	} else {
		tmPkgExternal = tmPkgLocal
	}

	tmPkgLocal[tLocal.String()] = tExternal
	tmPkgLocal[reflect.PtrTo(tLocal).String()] = reflect.PtrTo(tExternal)

	if tLocal != tExternal {
		tmPkgExternal[tExternal.String()] = tLocal
		tmPkgExternal[reflect.PtrTo(tExternal).String()] = reflect.PtrTo(tLocal)
		tr.All[tExternal.String()] = tExternal
		tr.All[reflect.PtrTo(tExternal).String()] = reflect.PtrTo(tExternal)
	}
}

func (tr *typeRegistry) Translate(x interface{}) interface{} {
	// fmt.Printf("Local: %v -> %v (%v)\n", x, reflect.TypeOf(x).String(), byPackage)
	v := reflect.ValueOf(x)
	t := v.Type()
	if target, ok := tr.ByPackage[t.PkgPath()][t.String()]; ok {
		return v.Convert(target).Interface()
	}
	fmt.Printf("Warning: tried to convert %v to local alias, but couldn't find it in the type map %v\n", reflect.TypeOf(x).String(), tr.ByPackage)
	return x
}

func Translate(x interface{}) interface{} {
	return Instance().Translate(x)
}

/* func LocalToExternal(x interface{}, reflect.Type target) interface{} {
	v := reflect.ValueOf(x)
	t := v.Type()
} */
