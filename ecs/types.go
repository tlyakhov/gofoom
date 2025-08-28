// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"tlyakhov/gofoom/containers"

	"github.com/traefik/yaegi/interp"
)

type ComponentID uint32
type controllerMetadata struct {
	Constructor func() Controller
	Type        reflect.Type
	Priority    int
}
type typeMetadata struct {
	ArenaIndexes         map[string]int
	IDs                  map[string]ComponentID
	LenGroupedComponents int
	ArenaPlaceholders    []AttachableArena
	nextFreeComponent    uint32
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
			Controllers:  make([]controllerMetadata, 0),
			IDs:          make(map[string]ComponentID),
			ArenaIndexes: make(map[string]int),
			ExprEnv:      make(map[string]any),
		}
	})
	return globalTypeMetadata
}

func RegisterComponent[T any, PT GenericAttachable[T]](arena *Arena[T, PT]) ComponentID {
	ecsTypes := Types()
	ecsTypes.lock.Lock()
	defer ecsTypes.lock.Unlock()
	arenaIndex := (int)(atomic.AddUint32(&ecsTypes.nextFreeComponent, 1))
	for len(ecsTypes.ArenaPlaceholders) < arenaIndex+1 {
		ecsTypes.ArenaPlaceholders = append(ecsTypes.ArenaPlaceholders, nil)
	}
	// Remove the package prefix (e.g. "core.Body" -> "Body")
	// see core.Expression
	arena.typeOfT = reflect.TypeFor[T]()
	tSplit := strings.Split(arena.Type().String(), ".")
	noPackage := tSplit[len(tSplit)-1]
	cid := ComponentID(arenaIndex)
	arena.componentID = cid
	arena.Getter = func(e Entity) PT {
		if asserted, ok := Component(e, cid).(PT); ok {
			return asserted
		}
		return nil
	}
	arena.isSingleton = false
	if sf, ok := arena.typeOfT.FieldByName("Attached"); ok {
		if sf.Tag.Get("ecs") == "singleton" {
			arena.isSingleton = true
		}
	}

	ecsTypes.ArenaPlaceholders[arenaIndex] = arena
	ecsTypes.ArenaIndexes[reflect.PointerTo(arena.Type()).String()] = arenaIndex
	ecsTypes.ArenaIndexes[arena.String()] = arenaIndex
	ecsTypes.IDs[reflect.PointerTo(arena.Type()).String()] = arena.componentID
	ecsTypes.IDs[arena.String()] = arena.componentID
	ecsTypes.ExprEnv[noPackage] = arena.Getter
	return arena.componentID
}

func (ecsTypes *typeMetadata) Type(name string) AttachableArena {
	if index, ok := ecsTypes.ArenaIndexes[name]; ok {
		return ecsTypes.ArenaPlaceholders[index]
	}
	return nil
}

func SerializeComponentIDs(ids containers.Set[ComponentID]) string {
	s := ""
	for id := range ids {
		if len(s) > 0 {
			s += ","
		}
		s += Types().ArenaPlaceholders[id].String()
	}
	return s
}

func ParseComponentIDs(ids string) containers.Set[ComponentID] {
	result := make(containers.Set[ComponentID])
	split := strings.Split(ids, ",")
	for _, s := range split {
		id := Types().IDs[strings.Trim(s, " \t\r\n")]
		if id != 0 {
			result.Add(id)
		}
	}
	return result
}
