package concepts

import (
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
)

type dbTypes struct {
	Indexes           map[string]int
	Types             []reflect.Type
	Funcs             []any
	nextFreeComponent uint32
	Controllers       map[string]reflect.Type
	ExprEnv           map[string]any
	lock              sync.RWMutex
}

var globalDbTypes *dbTypes
var once sync.Once

func DbTypes() *dbTypes {
	once.Do(func() {
		globalDbTypes = &dbTypes{
			Controllers: make(map[string]reflect.Type),
			Indexes:     make(map[string]int),
			ExprEnv:     make(map[string]any),
		}
	})
	return globalDbTypes
}

func (db *dbTypes) Register(local any, fromDbFunc any) int {
	db.lock.Lock()
	defer db.lock.Unlock()
	tLocal := reflect.ValueOf(local).Type()

	if tLocal.Kind() == reflect.Ptr {
		tLocal = tLocal.Elem()
	}
	index := (int)(atomic.AddUint32(&db.nextFreeComponent, 1))
	for len(db.Types) < index+1 {
		db.Types = append(db.Types, nil)
		db.Funcs = append(db.Funcs, nil)
	}
	db.Types[index] = tLocal
	db.Funcs[index] = fromDbFunc
	db.Indexes[reflect.PtrTo(tLocal).String()] = index
	db.Indexes[tLocal.String()] = index

	tSplit := strings.Split(tLocal.String(), ".")
	db.ExprEnv[tSplit[len(tSplit)-1]] = fromDbFunc
	return index
}

func (db *dbTypes) Type(name string) reflect.Type {
	if index, ok := db.Indexes[name]; ok {
		return db.Types[index]
	}
	return nil
}
