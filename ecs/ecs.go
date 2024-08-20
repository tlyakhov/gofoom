// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"sync"

	"github.com/kelindar/bitmap"
)

// The architecture is like this:
// * An entity is a globally unique integer (uint32), e.g. primary key
// * A component is a named (string) table with columns of fields, rows of
// entities
// * A system is code that queries and operates on components and entities
type ECS struct {
	*Simulation
	EntityComponents [][]Attachable

	components   [][]Attachable
	usedEntities bitmap.Bitmap
	Lock         sync.RWMutex
}

func NewECS() *ECS {
	db := &ECS{}
	db.Clear()

	return db
}

func (db *ECS) Clear() {
	db.usedEntities = bitmap.Bitmap{}
	db.usedEntities.Set(0) // 0 is reserved
	db.EntityComponents = make([][]Attachable, 1)
	db.components = make([][]Attachable, len(Types().Types))
	db.Simulation = NewSimulation()
	for i := 0; i < len(Types().Types); i++ {
		db.components[i] = make([]Attachable, 0)
	}
}

// Reserves an entity ID in the database (no components attached)
func (db *ECS) NewEntity() Entity {
	if free, found := db.usedEntities.MinZero(); found {
		db.usedEntities.Set(free)
		return Entity(free)
	}
	nextFree := len(db.EntityComponents)
	db.EntityComponents = append(db.EntityComponents, nil)
	db.usedEntities.Set(uint32(nextFree))
	return Entity(nextFree)
}

func (db *ECS) AllOfType(index int) []Attachable {
	return db.components[index]
}

func (db *ECS) AllComponents(entity Entity) []Attachable {
	if entity == 0 || len(db.EntityComponents) <= int(entity) {
		return nil
	}
	return db.EntityComponents[entity]
}

// Callers need to be careful, this function can return nil that's not castable
// to an actual component type. The *FromDb methods are better.
func (db *ECS) Component(entity Entity, index int) Attachable {
	if entity == 0 || index == 0 || len(db.EntityComponents) <= int(entity) {
		return nil
	}
	ec := db.EntityComponents[entity]
	if ec == nil {
		return nil
	}
	return ec[index]
}

func (db *ECS) AllOfNamedType(cType string) []Attachable {
	if index, ok := Types().Indexes[cType]; ok {
		return db.components[index]
	}
	return nil
}

func (db *ECS) First(index int) Attachable {
	for _, c := range db.components[index] {
		return c
	}
	return nil
}

// Attach a component to an entity. If a component with this type is already
// attached, this method will overwrite it.
func (db *ECS) attach(entity Entity, component Attachable, index int) {
	if entity == 0 {
		log.Printf("Tried to attach 0 entity!")
		return
	}
	component.SetEntity(entity)
	component.SetECS(db)

	for len(db.EntityComponents) <= int(entity) {
		db.EntityComponents = append(db.EntityComponents, nil)
	}

	var ec []Attachable
	if ec = db.EntityComponents[entity]; ec != nil {
		if ec[index] != nil {
			// A component with this index is already attached to this entity, overwrite it.
			componentsIndex := ec[index].IndexInECS()
			component.SetIndexInECS(componentsIndex)
			db.components[index][componentsIndex] = component
			ec[index] = component
			return
		}
	} else {
		ec = make([]Attachable, len(Types().Types))
		db.EntityComponents[entity] = ec
	}
	// This entity doesn't have a component with this index attached. Extend the
	// slice.
	ec[index] = component
	db.components[index] = append(db.components[index], component)
	component.SetIndexInECS(len(db.components[index]) - 1)
}

// Create a new component with the given index and attach it.
func (db *ECS) NewAttachedComponent(entity Entity, index int) Attachable {
	t := Types().Types[index]
	newc := reflect.New(t).Interface()
	attached := newc.(Attachable)
	db.attach(entity, attached, index)
	attached.Construct(nil)
	return attached
}

func (db *ECS) LoadAttachComponent(index int, data map[string]any, ignoreSerializedEntity bool) Attachable {
	var entity Entity
	var err error
	if ignoreSerializedEntity || data["Entity"] == nil {
		entity = db.NewEntity()
	} else if entity, err = ParseEntity(data["Entity"].(string)); entity == 0 || err != nil {
		entity = db.NewEntity()
	}

	db.usedEntities.Set(uint32(entity))

	t := Types().Types[index]
	newc := reflect.New(t).Interface()
	attached := newc.(Attachable)
	db.attach(entity, attached, index)
	attached.Construct(data)
	return attached
}

func (db *ECS) LoadComponentWithoutAttaching(index int, data map[string]any) Attachable {
	if data == nil {
		return nil
	}
	t := Types().Types[index]
	newc := reflect.New(t).Interface()
	component := newc.(Attachable)
	component.SetEntity(0)
	component.SetECS(db)
	component.Construct(data)
	return component
}

func (db *ECS) NewAttachedComponentTyped(entity Entity, cType string) Attachable {
	if index, ok := Types().Indexes[cType]; ok {
		return db.NewAttachedComponent(entity, index)
	}

	log.Printf("NewComponent: unregistered type %v for entity %v\n", cType, entity)
	return nil
}

func (db *ECS) Attach(componentIndex int, entity Entity, component Attachable) {
	db.attach(entity, component, componentIndex)
}

// This seems expensive. Need to profile
func (db *ECS) AttachTyped(entity Entity, component Attachable) {
	t := reflect.ValueOf(component).Type()

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if index, ok := Types().Indexes[t.String()]; ok {
		db.attach(entity, component, index)
	}
}

func (db *ECS) detach(index int, entity Entity, checkForEmpty bool) {
	if entity == 0 || index == 0 {
		log.Printf("ECS.Detach: tried to detach 0 entity/index.")
		return
	}

	if len(db.EntityComponents) <= int(entity) {
		log.Printf("ECS.Detach: entity %v is >= length of list %v.", entity, len(db.EntityComponents))
		return
	}
	ec := db.EntityComponents[entity]
	if ec == nil {
		log.Printf("ECS.Detach: entity %v has no components.", entity)
		return
	}

	if ec[index] == nil {
		// log.Printf("ECS.Detach: entity %v component %v is nil.", entity, index)
		return
	}
	ec[index].OnDetach()
	i := ec[index].IndexInECS()
	components := db.components[index]
	size := len(components)
	if size > i {
		components[i] = components[size-1]
		components[i].SetIndexInECS(i)
		db.components[index] = components[:size-1]
	} else {
		log.Printf("ECS.Detach: found entity %v component index %v, but component list is too short.", entity, index)
	}
	ec[index] = nil

	if checkForEmpty {
		allNil := true
		for _, c := range ec {
			if c != nil {
				allNil = false
			}
		}
		if allNil {
			db.EntityComponents[entity] = nil
			db.usedEntities.Remove(uint32(entity))
		}
	}
}

func (db *ECS) Detach(index int, entity Entity) {
	db.detach(index, entity, true)
}

func (db *ECS) DetachByType(component Attachable) {
	if component == nil {
		return
	}
	entity := component.GetEntity()

	if entity == 0 {
		return
	}

	tLocal := reflect.ValueOf(component).Type()

	if tLocal.Kind() == reflect.Ptr {
		tLocal = tLocal.Elem()
	}

	index := Types().Indexes[tLocal.String()]
	if index < 0 {
		return
	}

	db.detach(index, entity, true)
}

func (db *ECS) DetachAll(entity Entity) {
	if entity == 0 {
		return
	}

	for index := range db.components {
		db.detach(index, entity, false)
	}

	db.EntityComponents[entity] = nil
	db.usedEntities.Remove(uint32(entity))
}

func (db *ECS) GetEntityByName(name string) Entity {
	if allNamed := db.AllOfType(NamedComponentIndex); allNamed != nil {
		for _, c := range allNamed {
			named := c.(*Named)
			if named.Name == name {
				return named.Entity
			}
		}
	}
	return 0
}

func (db *ECS) DeserializeAndAttachEntity(jsonEntity map[string]any) {
	for name, index := range Types().Indexes {
		jsonData := jsonEntity[name]
		if jsonData == nil {
			continue
		}
		jsonComponent := jsonData.(map[string]any)
		db.LoadAttachComponent(index, jsonComponent, false)
	}
}
func (db *ECS) Load(filename string) error {
	db.Lock.Lock()
	defer db.Lock.Unlock()

	// TODO: Streaming loads?
	fileContents, err := os.ReadFile(filename)

	if err != nil {
		return err
	}

	var parsed any
	err = json.Unmarshal(fileContents, &parsed)
	if err != nil {
		return err
	}

	var jsonEntities []any
	var ok bool
	if jsonEntities, ok = parsed.([]any); !ok || jsonEntities == nil {
		return fmt.Errorf("ECS JSON root must be an array")
	}

	for _, jsonData := range jsonEntities {
		jsonEntity := jsonData.(map[string]any)
		if jsonEntity == nil {
			log.Printf("ECS JSON array element should be an object\n")
			continue
		}
		db.DeserializeAndAttachEntity(jsonEntity)
	}

	// After everything's loaded, trigger the controllers
	db.ActAllControllers(ControllerRecalculate)
	db.ActAllControllers(ControllerLoaded)
	return nil
}

func (db *ECS) SerializeEntity(entity Entity) map[string]any {
	components := db.EntityComponents[entity]
	jsonEntity := make(map[string]any)
	jsonEntity["Entity"] = entity.String()
	for index, component := range components {
		if component == nil || component.IsSystem() {
			continue
		}
		jsonEntity[Types().Types[index].String()] = component.Serialize()
	}
	if len(jsonEntity) == 1 {
		return nil
	}
	return jsonEntity
}

func (db *ECS) Save(filename string) {
	db.Lock.Lock()
	defer db.Lock.Unlock()
	jsonECS := make([]any, 0)

	for entity := range db.EntityComponents {
		jsonEntity := db.SerializeEntity(Entity(entity))
		if len(jsonEntity) == 0 {
			continue
		}
		jsonECS = append(jsonECS, jsonEntity)
	}

	bytes, err := json.MarshalIndent(jsonECS, "", "  ")

	if err != nil {
		panic(err)
	}

	os.WriteFile(filename, bytes, os.ModePerm)
}

func ConstructArray(parent any, arrayPtr any, data any) {
	valuePtr := reflect.ValueOf(arrayPtr)
	arrayValue := valuePtr.Elem()

	itemType := reflect.TypeOf(arrayPtr).Elem().Elem()
	arrayValue.Set(reflect.Zero(arrayValue.Type()))
	for _, child := range data.([]any) {
		item := reflect.New(itemType.Elem()).Interface().(Attachable)
		//item.SetParent(parent)
		item.Construct(child.(map[string]any))
		arrayValue.Set(reflect.Append(arrayValue, reflect.ValueOf(item)))
	}
}
