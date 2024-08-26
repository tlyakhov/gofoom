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
	ComponentColumns  []AttachableColumn
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

func RegisterComponent[T any, PT GenericAttachable[T]](column *Column[T, PT]) int {
	ecsTypes := Types()
	ecsTypes.lock.Lock()
	defer ecsTypes.lock.Unlock()
	index := (int)(atomic.AddUint32(&ecsTypes.nextFreeComponent, 1))
	for len(ecsTypes.ComponentColumns) < index+1 {
		ecsTypes.ComponentColumns = append(ecsTypes.ComponentColumns, nil)
	}
	// Remove the package prefix (e.g. "core.Body" -> "Body")
	// see core.Expression
	column.typeOfT = reflect.TypeFor[T]()
	tSplit := strings.Split(column.Type().String(), ".")
	noPackage := tSplit[len(tSplit)-1]

	ecsTypes.ComponentColumns[index] = column
	ecsTypes.Indexes[reflect.PointerTo(column.Type()).String()] = index
	ecsTypes.Indexes[column.String()] = index
	ecsTypes.IndexesNoPackage[noPackage] = index
	ecsTypes.ExprEnv[noPackage] = column.Getter
	return index
}

func (ecsTypes *typeMetadata) Type(name string) AttachableColumn {
	if index, ok := ecsTypes.Indexes[name]; ok {
		return ecsTypes.ComponentColumns[index]
	}
	return nil
}
