// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package concepts

import (
	"reflect"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/traefik/yaegi/interp"
)

type dbTypes struct {
	Indexes           map[string]int
	IndexesNoPackage  map[string]int
	Types             []reflect.Type
	Funcs             []any
	nextFreeComponent uint32
	Controllers       []reflect.Type
	ExprEnv           map[string]any
	InterpSymbols     interp.Exports
	lock              sync.RWMutex
}

var globalDbTypes *dbTypes
var once sync.Once

func DbTypes() *dbTypes {
	once.Do(func() {
		globalDbTypes = &dbTypes{
			Controllers:      make([]reflect.Type, 0),
			Indexes:          make(map[string]int),
			IndexesNoPackage: make(map[string]int),
			ExprEnv:          make(map[string]any),
		}
	})
	return globalDbTypes
}

func (dbt *dbTypes) Register(local any, fromDbFunc any) int {
	dbt.lock.Lock()
	defer dbt.lock.Unlock()
	tLocal := reflect.ValueOf(local).Type()

	if tLocal.Kind() == reflect.Ptr {
		tLocal = tLocal.Elem()
	}
	index := (int)(atomic.AddUint32(&dbt.nextFreeComponent, 1))
	for len(dbt.Types) < index+1 {
		dbt.Types = append(dbt.Types, nil)
		dbt.Funcs = append(dbt.Funcs, nil)
	}
	// Remove the package prefix (e.g. "core.Body" -> "Body")
	// see core.Expression
	tSplit := strings.Split(tLocal.String(), ".")
	noPackage := tSplit[len(tSplit)-1]

	dbt.Types[index] = tLocal
	dbt.Funcs[index] = fromDbFunc
	dbt.Indexes[reflect.PtrTo(tLocal).String()] = index
	dbt.Indexes[tLocal.String()] = index
	dbt.IndexesNoPackage[noPackage] = index
	dbt.ExprEnv[noPackage] = fromDbFunc
	camelCase := strings.ToLower(noPackage[0:1]) + noPackage[1:]
	dbt.ExprEnv["i"+camelCase] = (*EntityRef)(nil)
	return index
}

func (dbt *dbTypes) Type(name string) reflect.Type {
	if index, ok := dbt.Indexes[name]; ok {
		return dbt.Types[index]
	}
	return nil
}
