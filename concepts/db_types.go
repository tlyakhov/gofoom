package concepts

import (
	"reflect"
	"sync"
	"sync/atomic"
)

type ControllerMetadata struct {
	reflect.Type
	Methods map[string]reflect.Value
}

type dbTypes struct {
	Indexes           map[string]int
	Types             []reflect.Type
	nextFreeComponent uint32
	Controllers       map[string]*ControllerMetadata
	lock              sync.RWMutex
}

var globalDbTypes *dbTypes
var once sync.Once

func DbTypes() *dbTypes {
	once.Do(func() {
		globalDbTypes = &dbTypes{
			Controllers: make(map[string]*ControllerMetadata),
			Indexes:     make(map[string]int),
		}
	})
	return globalDbTypes
}

func (db *dbTypes) Register(local any) int {
	db.lock.Lock()
	defer db.lock.Unlock()
	tLocal := reflect.ValueOf(local).Type()

	if tLocal.Kind() == reflect.Ptr {
		tLocal = tLocal.Elem()
	}
	index := (int)(atomic.AddUint32(&db.nextFreeComponent, 1))
	db.Types = db.Types[:index+1]
	db.Types[index] = tLocal
	db.Indexes[reflect.PtrTo(tLocal).String()] = index
	db.Indexes[tLocal.String()] = index
	return index
}

func (db *dbTypes) Type(name string) reflect.Type {
	if index, ok := db.Indexes[name]; ok {
		return db.Types[index]
	}
	return nil
}
