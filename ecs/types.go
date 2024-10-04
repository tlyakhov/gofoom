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

type ComponentID uint32
type controllerMetadata struct {
	Constructor func() Controller
	Type        reflect.Type
	Priority    int
}
type typeMetadata struct {
	ColumnIndexes        map[string]int
	GroupIndexes         map[string]int
	IDs                  map[string]ComponentID
	LenGroupedComponents int
	ColumnPlaceholders   []AttachableColumn
	nextFreeComponent    uint32
	nextFreeGroup        uint32
	Controllers          []controllerMetadata
	ExprEnv              map[string]any
	InterpSymbols        interp.Exports
	lock                 sync.RWMutex
}

var globalTypeMetadata *typeMetadata
var once sync.Once

func Types() *typeMetadata {
	once.Do(func() {
		globalTypeMetadata = &typeMetadata{
			Controllers:   make([]controllerMetadata, 0),
			IDs:           make(map[string]ComponentID),
			ColumnIndexes: make(map[string]int),
			GroupIndexes:  make(map[string]int),
			ExprEnv:       make(map[string]any),
		}
	})
	return globalTypeMetadata
}

func RegisterComponent[T any, PT GenericAttachable[T]](column *Column[T, PT], group string) ComponentID {
	ecsTypes := Types()
	ecsTypes.lock.Lock()
	defer ecsTypes.lock.Unlock()
	columnIndex := (int)(atomic.AddUint32(&ecsTypes.nextFreeComponent, 1))
	for len(ecsTypes.ColumnPlaceholders) < columnIndex+1 {
		ecsTypes.ColumnPlaceholders = append(ecsTypes.ColumnPlaceholders, nil)
	}
	// Remove the package prefix (e.g. "core.Body" -> "Body")
	// see core.Expression
	column.typeOfT = reflect.TypeFor[T]()
	tSplit := strings.Split(column.Type().String(), ".")
	noPackage := tSplit[len(tSplit)-1]

	var groupIndex int
	switch {
	case group == "":
		groupIndex = (int)(atomic.AddUint32(&ecsTypes.nextFreeGroup, 1))
		ecsTypes.LenGroupedComponents++
	case ecsTypes.GroupIndexes[group] != 0:
		groupIndex = ecsTypes.GroupIndexes[group]
	default:
		groupIndex = (int)(atomic.AddUint32(&ecsTypes.nextFreeGroup, 1))
		ecsTypes.GroupIndexes[group] = groupIndex
		ecsTypes.LenGroupedComponents++
	}

	column.componentID = ComponentID(groupIndex<<16 | columnIndex)
	ecsTypes.ColumnPlaceholders[columnIndex] = column
	ecsTypes.ColumnIndexes[reflect.PointerTo(column.Type()).String()] = columnIndex
	ecsTypes.ColumnIndexes[column.String()] = columnIndex
	ecsTypes.IDs[reflect.PointerTo(column.Type()).String()] = column.componentID
	ecsTypes.IDs[column.String()] = column.componentID
	ecsTypes.ExprEnv[noPackage] = column.Getter
	return column.componentID
}

func (ecsTypes *typeMetadata) Type(name string) AttachableColumn {
	if index, ok := ecsTypes.ColumnIndexes[name]; ok {
		return ecsTypes.ColumnPlaceholders[index]
	}
	return nil
}

// TODO: Benchmark this, seems painful
func (ecsTypes *typeMetadata) ID(c Attachable) ComponentID {
	name := reflect.TypeOf(c).Elem().String()
	if id, ok := ecsTypes.IDs[name]; ok {
		return id
	}
	return 0
}
