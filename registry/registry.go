package registry

import (
	"fmt"
	"path"
	"reflect"
	"sync"
)

type typeRegistry struct {
	ByPackage map[string]map[reflect.Type]reflect.Type
	Inverse   map[string]map[reflect.Type]reflect.Type
	All       map[string]reflect.Type
}

var instance *typeRegistry
var once sync.Once

func Instance() *typeRegistry {
	once.Do(func() {
		instance = &typeRegistry{
			ByPackage: make(map[string]map[reflect.Type]reflect.Type),
			Inverse:   make(map[string]map[reflect.Type]reflect.Type),
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

	pkgLocal := path.Base(tLocal.PkgPath())
	pkgExternal := path.Base(tExternal.PkgPath())

	var tmPkgLocal map[reflect.Type]reflect.Type
	var tmPkgInverse map[reflect.Type]reflect.Type
	var ok bool

	if tmPkgLocal, ok = tr.ByPackage[pkgLocal]; !ok {
		tmPkgLocal = make(map[reflect.Type]reflect.Type)
		tr.ByPackage[pkgLocal] = tmPkgLocal
	}

	if pkgLocal != pkgExternal {
		if tmPkgInverse, ok = tr.Inverse[pkgExternal]; !ok {
			tmPkgInverse = make(map[reflect.Type]reflect.Type)
			tr.Inverse[pkgExternal] = tmPkgInverse
		}
	}

	tmPkgLocal[tExternal] = tLocal
	tmPkgLocal[reflect.PtrTo(tExternal)] = reflect.PtrTo(tLocal)

	if tLocal != tExternal {
		//tmPkgLocal[tExternal.String()] = tLocal
		//tmPkgLocal[reflect.PtrTo(tExternal).String()] = reflect.PtrTo(tLocal)
		tmPkgInverse[tLocal] = tExternal
		tmPkgInverse[reflect.PtrTo(tLocal)] = reflect.PtrTo(tExternal)
		tr.All[tExternal.String()] = tExternal
		tr.All[reflect.PtrTo(tExternal).String()] = reflect.PtrTo(tExternal)
	}
}

func (tr *typeRegistry) Translate(x interface{}, pkg string) interface{} {
	// fmt.Printf("Local: %v -> %v (%v)\n", x, reflect.TypeOf(x).String(), byPackage)
	t := reflect.TypeOf(x)
	if target, ok := tr.ByPackage[pkg][t]; ok {
		v := reflect.ValueOf(x)
		return v.Convert(target).Interface()
	} else if target, ok := tr.Inverse[pkg][t]; ok {
		v := reflect.ValueOf(x)
		return v.Convert(target).Interface()
	}
	fmt.Printf("Warning: tried to convert %v to %v alias, but couldn't find it in the registry.\n", reflect.TypeOf(x).String(), pkg)
	return x
}

func Translate(x interface{}, pkg string) interface{} {
	return Instance().Translate(x, pkg)
}

/* func LocalToExternal(x interface{}, reflect.Type target) interface{} {
	v := reflect.ValueOf(x)
	t := v.Type()
} */
