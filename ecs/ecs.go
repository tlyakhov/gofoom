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
	Entities bitmap.Bitmap
	Lock     sync.RWMutex

	rows    []ComponentTable
	columns []AttachableColumn
}

func NewECS() *ECS {
	db := &ECS{}
	db.Clear()

	return db
}

func (db *ECS) Clear() {
	db.Entities = bitmap.Bitmap{}
	// 0 is reserved
	db.Entities.Set(0)
	// 0 is reserved
	db.rows = make([]ComponentTable, 1)
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
	nextFree := len(db.rows)
	for len(db.rows) < (nextFree + 1) {
		db.rows = append(db.rows, nil)
	}
	db.Entities.Set(uint32(nextFree))
	return Entity(nextFree)
}

func ColumnFor[T any, PT GenericAttachable[T]](db *ECS, id ComponentID) *Column[T, PT] {
	return db.columns[id].(*Column[T, PT])
}

func (db *ECS) Column(id ComponentID) AttachableColumn {
	return db.columns[id]
}

func (db *ECS) AllComponents(entity Entity) ComponentTable {
	if entity == 0 || len(db.rows) <= int(entity) {
		return nil
	}
	return db.rows[int(entity)]
}

// Callers need to be careful, this function can return nil that's not castable
// to an actual component type. The Get* methods are better.
func (db *ECS) Component(entity Entity, id ComponentID) Attachable {
	if entity == 0 || id == 0 || len(db.rows) <= int(entity) {
		return nil
	}
	return db.rows[int(entity)].Get(id)
}

// Shortcut methods for retrieving multiple components at a time
func Archetype2[T1 any, T2 any,
	PT1 GenericAttachable[T1], PT2 GenericAttachable[T2]](db *ECS, entity Entity) (pt1 PT1, pt2 PT2) {
	pt1 = nil
	pt2 = nil
	if entity == 0 || len(db.rows) <= int(entity) {
		return
	}
	for _, attachable := range db.rows[int(entity)] {
		if attachable == nil {
			continue
		}
		if ct, ok := attachable.(PT1); ok {
			pt1 = ct
		} else if ct, ok := attachable.(PT2); ok {
			pt2 = ct
		}
		if pt1 != nil && pt2 != nil {
			return
		}
	}
	return
}

func Archetype3[T1 any, T2 any, T3 any,
	PT1 GenericAttachable[T1],
	PT2 GenericAttachable[T2],
	PT3 GenericAttachable[T3]](db *ECS, entity Entity) (pt1 PT1, pt2 PT2, pt3 PT3) {
	pt1 = nil
	pt2 = nil
	pt3 = nil
	if entity == 0 || len(db.rows) <= int(entity) {
		return
	}
	for _, attachable := range db.rows[int(entity)] {
		if attachable == nil {
			continue
		}
		if ct, ok := attachable.(PT1); ok {
			pt1 = ct
		} else if ct, ok := attachable.(PT2); ok {
			pt2 = ct
		} else if ct, ok := attachable.(PT3); ok {
			pt3 = ct
		}

		if pt1 != nil && pt2 != nil && pt3 != nil {
			return
		}
	}
	return
}

func (db *ECS) First(id ComponentID) Attachable {
	c := db.columns[id]
	for i := 0; i < c.Cap(); i++ {
		a := db.columns[id].Attachable(i)
		if a != nil {
			return a
		}
	}
	return nil
}

// Attach a component to an entity. If a component with this type is already
// attached, this method will overwrite it.
func (db *ECS) upsert(entity Entity, component Attachable, componentID ComponentID) Attachable {
	if entity == 0 {
		log.Printf("Tried to attach 0 entity!")
		return nil
	}

	for int(entity) >= len(db.rows) {
		db.rows = append(db.rows, nil)
	}

	column := db.columns[componentID]
	ec := db.rows[int(entity)].Get(componentID)
	if ec != nil {
		// A component with this index is already attached to this entity, overwrite it.
		indexInColumn := ec.Base().indexInColumn
		component = column.Replace(component, indexInColumn)
	} else {
		// This entity doesn't have a component with this index attached. Extend the
		// slice.
		component = column.Add(component)
	}
	component.Base().Entity = entity
	component.Base().ComponentID = componentID
	db.rows[int(entity)].Set(component)
	return component
}

// Create a new component with the given index and attach it.
func (db *ECS) NewAttachedComponent(entity Entity, id ComponentID) Attachable {
	attached := db.upsert(entity, nil, id)
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

	attached := db.upsert(entity, nil, id)
	attached.Construct(data)
	return attached
}

func (db *ECS) LoadComponentWithoutAttaching(id ComponentID, data map[string]any) Attachable {
	if data == nil {
		return nil
	}
	component := Types().ColumnPlaceholders[id].New()
	component.Base().Entity = 0
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

func (db *ECS) Upsert(id ComponentID, entity Entity, component Attachable) Attachable {
	return db.upsert(entity, component, id)
}

// This seems expensive. Need to profile
func (db *ECS) UpsertTyped(entity Entity, component Attachable) {
	t := reflect.ValueOf(component).Type()

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if index, ok := Types().IDs[t.String()]; ok {
		db.upsert(entity, component, index)
	}
}

func (db *ECS) delete(id ComponentID, entity Entity, checkForEmpty bool) {
	if entity == 0 {
		log.Printf("ECS.Detach: tried to detach 0 entity.")
		return
	}
	if id == 0 {
		log.Printf("ECS.Detach: tried to detach 0 component index.")
		return
	}

	if len(db.rows) <= int(entity) {
		log.Printf("ECS.Detach: entity %v is >= length of list %v.", entity, len(db.rows))
		return
	}
	ec := db.rows[int(entity)].Get(id)
	column := db.columns[id]
	if ec == nil {
		// This component is not attached
		log.Printf("ECS.Detach: tried to detach unattached component %v from entity %v", column.String(), entity)
		return
	}

	// Ensure that the component is of the type the caller expects:
	iic := ec.Base().indexInColumn
	if column.Len() <= iic || column.Attachable(iic) != ec {
		log.Printf("ECS.Detach: tried to detach %v from entity %v - %v", column.String(), entity, ec.String())
		return
	}

	ec.OnDetach()
	column.Detach(ec.Base().indexInColumn)
	db.rows[int(entity)].Delete(id)

	if checkForEmpty {
		allNil := true
		for _, a := range db.rows[int(entity)] {
			if a != nil {
				allNil = false
				break
			}
		}
		if allNil {
			db.Entities.Remove(uint32(entity))
			db.rows[int(entity)] = nil
		}
	}
}

func (db *ECS) DeleteComponent(id ComponentID, entity Entity) {
	db.delete(id, entity, true)
}

func (db *ECS) DeleteByType(component Attachable) {
	if component == nil {
		return
	}
	entity := component.Base().Entity

	if entity == 0 {
		return
	}

	tLocal := reflect.ValueOf(component).Type()

	if tLocal.Kind() == reflect.Ptr {
		tLocal = tLocal.Elem()
	}

	id := Types().IDs[tLocal.String()]
	db.delete(id, entity, true)
}

func (db *ECS) Delete(entity Entity) {
	if entity == 0 {
		return
	}

	if len(db.rows) <= int(entity) {
		return
	}

	for _, c := range db.rows[int(entity)] {
		if c == nil {
			continue
		}
		id := c.Base().ComponentID
		c.OnDetach()
		db.columns[id].Detach(c.Base().indexInColumn)
	}
	db.rows[int(entity)] = nil
	db.Entities.Remove(uint32(entity))
}

func (db *ECS) GetEntityByName(name string) Entity {
	col := ColumnFor[Named](db, NamedCID)
	for i := range col.Cap() {
		if named := col.Value(i); named != nil && named.Name == name {
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
	for _, component := range db.rows[int(entity)] {
		if component == nil || component.IsSystem() {
			continue
		}
		col := Types().ColumnPlaceholders[component.Base().ComponentID]
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
