// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package concepts

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"slices"
	"sync"

	"github.com/kelindar/bitmap"
)

// The architecture is like this:
// * An entity is a globally unique uint64, e.g. primary key
// * A component is a named (string) table with columns of fields, rows of
// entities
// * A system is code that queries and operates on components and entities
type EntityComponentDB struct {
	*Simulation
	EntityComponents [][]Attachable

	components   [][]Attachable
	usedEntities bitmap.Bitmap
	Lock         sync.RWMutex
}

func NewEntityComponentDB() *EntityComponentDB {
	db := &EntityComponentDB{}
	db.Clear()

	return db
}

func (db *EntityComponentDB) Clear() {
	db.usedEntities = bitmap.Bitmap{}
	db.usedEntities.Set(0) // 0 is reserved
	db.EntityComponents = make([][]Attachable, 1)
	db.components = make([][]Attachable, len(DbTypes().Types))
	db.Simulation = NewSimulation()
	for i := 0; i < len(DbTypes().Types); i++ {
		db.components[i] = make([]Attachable, 0)
	}
}

// Reserves an entity ID in the database (no components attached)
func (db *EntityComponentDB) NewEntity() Entity {
	if free, found := db.usedEntities.MinZero(); found {
		db.usedEntities.Set(free)
		return Entity(free)
	}
	nextFree := len(db.EntityComponents)
	db.EntityComponents = append(db.EntityComponents, nil)
	db.usedEntities.Set(uint32(nextFree))
	return Entity(nextFree)
}

func (db *EntityComponentDB) AllOfType(index int) []Attachable {
	return db.components[index]
}

func (db *EntityComponentDB) AllComponents(entity Entity) []Attachable {
	if entity == 0 || len(db.EntityComponents) <= int(entity) {
		return nil
	}
	return db.EntityComponents[entity]
}

// Callers need to be careful, this function can return nil that's not castable
// to an actual component type. The *FromDb methods are better.
func (db *EntityComponentDB) Component(entity Entity, index int) Attachable {
	if entity == 0 || index == 0 || len(db.EntityComponents) <= int(entity) {
		return nil
	}
	ec := db.EntityComponents[entity]
	if ec == nil {
		return nil
	}
	return ec[index]
}

func (db *EntityComponentDB) AllOfNamedType(cType string) []Attachable {
	if index, ok := DbTypes().Indexes[cType]; ok {
		return db.components[index]
	}
	return nil
}

func (db *EntityComponentDB) First(index int) Attachable {
	for _, c := range db.components[index] {
		return c
	}
	return nil
}

// Attach a component to an entity. If a component with this type is already
// attached, this method will overwrite it.
func (db *EntityComponentDB) attach(entity Entity, component Attachable, index int) {
	if entity == 0 {
		log.Printf("Tried to attach 0 entity!")
		return
	}
	component.SetEntity(entity)
	component.SetDB(db)

	for len(db.EntityComponents) <= int(entity) {
		db.EntityComponents = append(db.EntityComponents, nil)
	}

	var ec []Attachable
	if ec = db.EntityComponents[entity]; ec != nil {
		if ec[index] != nil {
			// A component with this index is already attached to this entity, overwrite it.
			componentsIndex := ec[index].IndexInDB()
			component.SetIndexInDB(componentsIndex)
			db.components[index][componentsIndex] = component
			ec[index] = component
			return
		}
	} else {
		ec = make([]Attachable, len(DbTypes().Types))
		db.EntityComponents[entity] = ec
	}
	// This entity doesn't have a component with this index attached. Extend the
	// slice.
	ec[index] = component
	db.components[index] = append(db.components[index], component)
	component.SetIndexInDB(len(db.components[index]) - 1)
}

// Create a new component with the given index and attach it.
func (db *EntityComponentDB) NewAttachedComponent(entity Entity, index int) Attachable {
	t := DbTypes().Types[index]
	newc := reflect.New(t).Interface()
	attached := newc.(Attachable)
	db.attach(entity, attached, index)
	attached.Construct(nil)
	return attached
}

func (db *EntityComponentDB) LoadAttachComponent(index int, data map[string]any, ignoreSerializedEntity bool) Attachable {
	var entity Entity
	var err error
	if ignoreSerializedEntity || data["Entity"] == nil {
		entity = db.NewEntity()
	} else if entity, err = ParseEntity(data["Entity"].(string)); entity == 0 || err != nil {
		entity = db.NewEntity()
	}

	db.usedEntities.Set(uint32(entity))

	t := DbTypes().Types[index]
	newc := reflect.New(t).Interface()
	attached := newc.(Attachable)
	db.attach(entity, attached, index)
	attached.Construct(data)
	return attached
}

func (db *EntityComponentDB) LoadComponentWithoutAttaching(index int, data map[string]any) Attachable {
	if data == nil {
		return nil
	}
	t := DbTypes().Types[index]
	newc := reflect.New(t).Interface()
	component := newc.(Attachable)
	component.SetEntity(0)
	component.SetDB(db)
	component.Construct(data)
	return component
}

func (db *EntityComponentDB) NewAttachedComponentTyped(entity Entity, cType string) Attachable {
	if index, ok := DbTypes().Indexes[cType]; ok {
		return db.NewAttachedComponent(entity, index)
	}

	log.Printf("NewComponent: unregistered type %v for entity %v\n", cType, entity)
	return nil
}

func (db *EntityComponentDB) Attach(componentIndex int, entity Entity, component Attachable) {
	db.attach(entity, component, componentIndex)
}

// This seems expensive. Need to profile
func (db *EntityComponentDB) AttachTyped(entity Entity, component Attachable) {
	t := reflect.ValueOf(component).Type()

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if index, ok := DbTypes().Indexes[t.String()]; ok {
		db.attach(entity, component, index)
	}
}

func (db *EntityComponentDB) Detach(index int, entity Entity) {
	if entity == 0 || index == 0 {
		log.Printf("EntityComponentDB.Detach: tried to detach 0 entity/index.")
		return
	}

	if len(db.EntityComponents) <= int(entity) {
		log.Printf("EntityComponentDB.Detach: entity %v is >= length of list %v.", entity, len(db.EntityComponents))
		return
	}
	ec := db.EntityComponents[entity]
	if ec == nil {
		log.Printf("EntityComponentDB.Detach: entity %v has no components.", entity)
		return
	}

	if ec[index] == nil {
		// log.Printf("EntityComponentDB.Detach: entity %v component %v is nil.", entity, index)
		return
	}
	i := ec[index].IndexInDB()
	components := db.components[index]
	size := len(components)
	if size > i {
		components[i] = components[size-1]
		components[i].SetIndexInDB(i)
		db.components[index] = components[:size-1]
	} else {
		log.Printf("EntityComponentDB.Detach: found entity %v component index %v, but component list is too short.", entity, index)
	}
	ec[index] = nil
}

func (db *EntityComponentDB) DetachByType(component Attachable) {
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

	index := DbTypes().Indexes[tLocal.String()]
	if index < 0 {
		return
	}

	db.Detach(index, entity)
	component.SetEntity(0)
}

func (db *EntityComponentDB) DetachAll(entity Entity) {
	if entity == 0 {
		return
	}

	for index := range db.components {
		db.Detach(index, entity)
	}

	db.EntityComponents[entity] = nil
}

func (db *EntityComponentDB) GetEntityByName(name string) Entity {
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

func (db *EntityComponentDB) DeserializeAndAttachEntity(jsonEntity map[string]any) {
	for name, index := range DbTypes().Indexes {
		jsonData := jsonEntity[name]
		if jsonData == nil {
			continue
		}
		jsonComponent := jsonData.(map[string]any)
		db.LoadAttachComponent(index, jsonComponent, false)
	}
}
func (db *EntityComponentDB) Load(filename string) error {
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

func (db *EntityComponentDB) SerializeEntity(entity Entity) map[string]any {
	components := db.EntityComponents[entity]
	jsonEntity := make(map[string]any)
	jsonEntity["Entity"] = entity.Format()
	for index, component := range components {
		if component == nil {
			continue
		}
		jsonEntity[DbTypes().Types[index].String()] = component.Serialize()
	}
	return jsonEntity
}

func (db *EntityComponentDB) Save(filename string) {
	db.Lock.Lock()
	defer db.Lock.Unlock()
	jsonDB := make([]any, 0)

	sortedEntities := make([]Entity, 0)
	for entity, c := range db.EntityComponents {
		if entity == 0 || c == nil {
			continue
		}
		sortedEntities = append(sortedEntities, Entity(entity))
	}
	slices.Sort(sortedEntities)

	for _, entity := range sortedEntities {
		jsonDB = append(jsonDB, db.SerializeEntity(entity))
	}

	bytes, err := json.MarshalIndent(jsonDB, "", "  ")

	if err != nil {
		panic(err)
	}

	os.WriteFile(filename, bytes, os.ModePerm)
}

func (db *EntityComponentDB) DeserializeEntities(data []any) []Entity {
	if data == nil {
		return nil
	}
	result := make([]Entity, len(data))

	for i, e := range data {
		if entity, err := ParseEntity(e.(string)); err == nil {
			result[i] = entity
		}
	}
	return result
}

func (db *EntityComponentDB) SerializeEntities(data []Entity) []string {
	result := make([]string, len(data))
	for i, e := range data {
		result[i] = e.Format()
	}
	return result
}
