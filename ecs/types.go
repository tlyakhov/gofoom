// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"reflect"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/traefik/yaegi/interp"
)

type controllerMetadata struct {
	Controller
	Type     reflect.Type
	Priority int
}
type typeMetadata struct {
	Indexes           map[string]int
	IndexesNoPackage  map[string]int
	Types             []reflect.Type
	Funcs             []any
	nextFreeComponent uint32
	Controllers       []controllerMetadata
	ExprEnv           map[string]any
	InterpSymbols     interp.Exports
	lock              sync.RWMutex
}

var globalTypeMetadata *typeMetadata
var once sync.Once

func Types() *typeMetadata {
	once.Do(func() {
		globalTypeMetadata = &typeMetadata{
			Controllers:      make([]controllerMetadata, 0),
			Indexes:          make(map[string]int),
			IndexesNoPackage: make(map[string]int),
			ExprEnv:          make(map[string]any),
		}
	})
	return globalTypeMetadata
}

func (ecsTypes *typeMetadata) Register(local any, fromDbFunc any) int {
	ecsTypes.lock.Lock()
	defer ecsTypes.lock.Unlock()
	tLocal := reflect.ValueOf(local).Type()

	if tLocal.Kind() == reflect.Ptr {
		tLocal = tLocal.Elem()
	}
	index := (int)(atomic.AddUint32(&ecsTypes.nextFreeComponent, 1))
	for len(ecsTypes.Types) < index+1 {
		ecsTypes.Types = append(ecsTypes.Types, nil)
		ecsTypes.Funcs = append(ecsTypes.Funcs, nil)
	}
	// Remove the package prefix (e.g. "core.Body" -> "Body")
	// see core.Expression
	tSplit := strings.Split(tLocal.String(), ".")
	noPackage := tSplit[len(tSplit)-1]

	ecsTypes.Types[index] = tLocal
	ecsTypes.Funcs[index] = fromDbFunc
	ecsTypes.Indexes[reflect.PointerTo(tLocal).String()] = index
	ecsTypes.Indexes[tLocal.String()] = index
	ecsTypes.IndexesNoPackage[noPackage] = index
	ecsTypes.ExprEnv[noPackage] = fromDbFunc
	return index
}

func (ecsTypes *typeMetadata) Type(name string) reflect.Type {
	if index, ok := ecsTypes.Indexes[name]; ok {
		return ecsTypes.Types[index]
	}
	return nil
}
