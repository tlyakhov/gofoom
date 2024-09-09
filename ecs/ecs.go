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
	"tlyakhov/gofoom/dynamic"

	"github.com/kelindar/bitmap"
)

// The architecture is like this:
// * An entity is a globally unique integer (uint32), e.g. primary key
// * An entity can be associated with multiple components (one of each kind)
// * A system (controller) is code that queries and operates on components and entities
type ECS struct {
	*dynamic.Simulation
	Entities         bitmap.Bitmap
	entityComponents []ComponentTable

	lenGroupedComponents int
	columns              []AttachableColumn
	Lock                 sync.RWMutex
}

func NewECS() *ECS {
	db := &ECS{}
	db.Clear()

	return db
}

func (db *ECS) Clear() {
	db.lenGroupedComponents = Types().LenGroupedComponents
	db.Entities = bitmap.Bitmap{}
	db.Entities.Set(0) // 0 is reserved
	// 0 is reserved
	db.entityComponents = make([]ComponentTable, 1)
	db.columns = make([]AttachableColumn, len(Types().ColumnPlaceholders))
	db.Simulation = dynamic.NewSimulation()
	for i, columnPlaceholder := range Types().ColumnPlaceholders {
		if columnPlaceholder == nil {
			continue
		}
		log.Printf("Component %v, index: %v", columnPlaceholder.Type().String(), i)
		// t = *ComponentColumn[T]
		t := reflect.TypeOf(columnPlaceholder)
		db.columns[i] = reflect.New(t.Elem()).Interface().(AttachableColumn)
		db.columns[i].From(columnPlaceholder, db)
	}
}

// Reserves an entity ID in the database (no components attached)
func (db *ECS) NewEntity() Entity {
	if free, found := db.Entities.MinZero(); found {
		db.Entities.Set(free)
		return Entity(free)
	}
	nextFree := len(db.entityComponents)
	for len(db.entityComponents) < (nextFree + 1) {
		db.entityComponents = append(db.entityComponents, nil)
	}
	db.Entities.Set(uint32(nextFree))
	return Entity(nextFree)
}

func ColumnFor[T any, PT GenericAttachable[T]](db *ECS, id ComponentID) *Column[T, PT] {
	return db.columns[id&0xFFFF].(*Column[T, PT])
}

func (db *ECS) AllComponents(entity Entity) ComponentTable {
	if entity == 0 || len(db.entityComponents) <= int(entity) {
		return nil
	}
	return db.entityComponents[int(entity)]
}

// Callers need to be careful, this function can return nil that's not castable
// to an actual component type. The Get* methods are better.
func (db *ECS) Component(entity Entity, id ComponentID) Attachable {
	if entity == 0 || id == 0 || len(db.entityComponents) <= int(entity) {
		return nil
	}
	return db.entityComponents[int(entity)].Get(id)
}

func (db *ECS) First(id ComponentID) Attachable {
	c := db.columns[id&0xFFFF]
	if c.Len() > 0 {
		return db.columns[id&0xFFFF].Attachable(0)
	}
	return nil
}

// Attach a component to an entity. If a component with this type is already
// attached, this method will overwrite it.
func (db *ECS) attach(entity Entity, component Attachable, componentID ComponentID) Attachable {
	if entity == 0 {
		log.Printf("Tried to attach 0 entity!")
		return nil
	}

	for int(entity) >= len(db.entityComponents) {
		db.entityComponents = append(db.entityComponents, nil)
	}

	column := db.columns[componentID&0xFFFF]
	ec := db.entityComponents[int(entity)].Get(componentID)
	if ec != nil {
		// A component with this index is already attached to this entity, overwrite it.
		indexInColumn := ec.IndexInColumn()
		component = column.Replace(component, indexInColumn)
	} else {
		// This entity doesn't have a component with this index attached. Extend the
		// slice.
		component = column.Add(component)
	}
	component.SetEntity(entity)
	component.SetComponentID(componentID)
	db.entityComponents[int(entity)].Set(component)
	return component
}

// Create a new component with the given index and attach it.
func (db *ECS) NewAttachedComponent(entity Entity, id ComponentID) Attachable {
	attached := db.attach(entity, nil, id)
	attached.Construct(nil)
	return attached
}

func (db *ECS) LoadAttachComponent(id ComponentID, data map[string]any, ignoreSerializedEntity bool) Attachable {
	var entity Entity
	var err error
	if ignoreSerializedEntity || data["Entity"] == nil {
		entity = db.NewEntity()
	} else if entity, err = ParseEntity(data["Entity"].(string)); entity == 0 || err != nil {
		entity = db.NewEntity()
	}

	db.Entities.Set(uint32(entity))

	attached := db.attach(entity, nil, id)
	attached.Construct(data)
	return attached
}

func (db *ECS) LoadComponentWithoutAttaching(id ComponentID, data map[string]any) Attachable {
	if data == nil {
		return nil
	}
	component := Types().ColumnPlaceholders[id&0xFFFF].New()
	component.SetEntity(0)
	component.AttachECS(db)
	component.Construct(data)
	return component
}

func (db *ECS) NewAttachedComponentTyped(entity Entity, cType string) Attachable {
	if index, ok := Types().IDs[cType]; ok {
		return db.NewAttachedComponent(entity, index)
	}

	log.Printf("NewComponent: unregistered type %v for entity %v\n", cType, entity)
	return nil
}

func (db *ECS) Attach(id ComponentID, entity Entity, component Attachable) Attachable {
	return db.attach(entity, component, id)
}

// This seems expensive. Need to profile
func (db *ECS) AttachTyped(entity Entity, component Attachable) {
	t := reflect.ValueOf(component).Type()

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if index, ok := Types().IDs[t.String()]; ok {
		db.attach(entity, component, index)
	}
}

func (db *ECS) detach(id ComponentID, entity Entity, checkForEmpty bool) {
	if entity == 0 {
		log.Printf("ECS.Detach: tried to detach 0 entity.")
		return
	}
	if id == 0 {
		log.Printf("ECS.Detach: tried to detach 0 component index.")
		return
	}

	if len(db.entityComponents) <= int(entity) {
		log.Printf("ECS.Detach: entity %v is >= length of list %v.", entity, len(db.entityComponents))
		return
	}
	ec := db.entityComponents[int(entity)].Get(id)
	column := db.columns[id&0xFFFF]
	if ec == nil {
		// This component is not attached
		log.Printf("ECS.Detach: tried to detach unattached component %v from entity %v", column.String(), entity)
		return
	}

	// Ensure that the component is of the type the caller expects:
	iic := ec.IndexInColumn()
	if column.Len() <= iic || column.Attachable(iic) != ec {
		log.Printf("ECS.Detach: tried to detach %v from entity %v - %v", column.String(), entity, ec.String())
		return
	}

	ec.OnDetach()
	column.Detach(ec.IndexInColumn())
	db.entityComponents[int(entity)].Delete(id)

	if checkForEmpty {
		allNil := true
		for _, a := range db.entityComponents[int(entity)] {
			if a != nil {
				allNil = false
				break
			}
		}
		if allNil {
			db.Entities.Remove(uint32(entity))
			db.entityComponents[int(entity)] = nil
		}
	}
}

func (db *ECS) Detach(id ComponentID, entity Entity) {
	db.detach(id, entity, true)
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

	id := Types().IDs[tLocal.String()]
	db.detach(id, entity, true)
}

func (db *ECS) DetachAll(entity Entity) {
	if entity == 0 {
		return
	}

	if len(db.entityComponents) <= int(entity) {
		return
	}

	for _, c := range db.entityComponents[int(entity)] {
		if c == nil {
			continue
		}
		id := c.GetComponentID()
		c.OnDetach()
		db.columns[id&0xFFFF].Detach(c.IndexInColumn())
	}
	db.entityComponents[int(entity)] = nil
	db.Entities.Remove(uint32(entity))
}

func (db *ECS) GetEntityByName(name string) Entity {
	col := ColumnFor[Named](db, NamedCID)
	for i := range col.Length {
		named := col.Value(i)
		if named.Name == name {
			return named.Entity
		}
	}

	return 0
}

func (db *ECS) DeserializeAndAttachEntity(jsonEntity map[string]any) {
	for name, index := range Types().IDs {
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
	jsonEntity := make(map[string]any)
	jsonEntity["Entity"] = entity.String()
	for _, component := range db.entityComponents[int(entity)] {
		if component == nil || component.IsSystem() {
			continue
		}
		col := Types().ColumnPlaceholders[component.GetComponentID()&0xFFFF]
		jsonEntity[col.Type().String()] = component.Serialize()
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

	db.Entities.Range(func(entity uint32) {
		jsonEntity := db.SerializeEntity(Entity(entity))
		if len(jsonEntity) == 0 {
			return
		}
		jsonECS = append(jsonECS, jsonEntity)
	})

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
