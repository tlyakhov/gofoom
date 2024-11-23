// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"sync"
	"tlyakhov/gofoom/dynamic"

	"github.com/kelindar/bitmap"
	"gopkg.in/yaml.v3"
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
		// log.Printf("Component %v, index: %v", columnPlaceholder.Type().String(), i)
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

func (db *ECS) Link(target Entity, source Entity) {
	if target == 0 || source == 0 ||
		len(db.rows) <= int(source) || len(db.rows) <= int(target) {
		return
	}
	for _, c := range db.rows[int(source)] {
		if c == nil || !c.MultiAttachable() {
			continue
		}
		db.attach(target, c, c.Base().ComponentID)
	}
}

// Attach a component to an entity. If a component with this type is already
// attached, this method will overwrite it.
func (db *ECS) attach(entity Entity, component Attachable, componentID ComponentID) Attachable {
	if entity == 0 {
		log.Printf("Tried to attach 0 entity!")
		return nil
	}

	for int(entity) >= len(db.rows) {
		db.rows = append(db.rows, nil)
	}

	// Try to retrieve the existing component for this entity
	ec := db.rows[int(entity)].Get(componentID)

	// Did the caller:
	// 1. not provide a component?
	// 2. the provided component is unattached?
	if component == nil || component.Base().Attachments == 0 {
		// Then we need to add a new element to the column:
		column := db.columns[componentID]
		if ec != nil {
			// A component with this index is already attached to this entity, overwrite it.
			indexInColumn := ec.Base().indexInColumn
			component = column.Replace(component, indexInColumn)
		} else {
			// This entity doesn't have a component with this index attached. Extend the
			// slice.
			component = column.Add(component)
		}
	} else if ec != nil {
		// We have a conflict between the provided component and an existing one
		// with the same component ID. We should abort.
		log.Printf("ECS.upsert: Entity %v already has a component %v. Aborting!", entity, Types().ColumnPlaceholders[componentID].String())
		return nil
	}

	a := component.Base()
	if a.Attachments > 0 && !component.MultiAttachable() {
		log.Printf("ECS.upsert: Component %v is already attached to %v and not multi-attachable.", component.String(), a.Entity)
	}
	a.Entities.Set(entity)
	a.Entity = entity
	a.Attachments++
	a.ComponentID = componentID
	db.rows[int(entity)].Set(component)
	return component
}

// Create a new component with the given index and attach it.
func (db *ECS) NewAttachedComponent(entity Entity, id ComponentID) Attachable {
	attached := db.attach(entity, nil, id)
	attached.Construct(nil)
	return attached
}

func (db *ECS) LoadAttachComponent(id ComponentID, data map[string]any, ignoreSerializedEntity bool) Attachable {
	entities, attachments := LoadEntitiesFromJson(data)
	if ignoreSerializedEntity || attachments == 0 {
		entities = make(EntityTable, 0)
		entities.Set(db.NewEntity())
	}

	var attached Attachable
	for _, entity := range entities {
		if entity == 0 {
			continue
		}
		db.Entities.Set(uint32(entity))
		attached = db.attach(entity, attached, id)
		if attached.Base().Attachments == 1 {
			attached.Construct(data)
		}
	}
	return attached
}

func (db *ECS) LoadComponentWithoutAttaching(id ComponentID, data map[string]any) Attachable {
	if data == nil {
		return nil
	}
	component := Types().ColumnPlaceholders[id].New()
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

func (db *ECS) AttachTyped(entity Entity, component Attachable) Attachable {
	return db.attach(entity, component, component.Base().ComponentID)
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

	ec.OnDetach(entity)
	if ec.Base().Attachments == 0 {
		column.Detach(ec.Base().indexInColumn)
		ec.OnDelete()
	}
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

func (db *ECS) DetachComponent(id ComponentID, entity Entity) {
	db.detach(id, entity, true)
}

func (db *ECS) DeleteByType(component Attachable) {
	if component == nil {
		return
	}
	if len(component.Base().Entities) == 0 {
		return
	}

	tLocal := reflect.ValueOf(component).Type()

	if tLocal.Kind() == reflect.Ptr {
		tLocal = tLocal.Elem()
	}

	id := Types().IDs[tLocal.String()]

	for _, entity := range component.Base().Entities {
		if entity != 0 {
			db.detach(id, entity, true)
		}
	}
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
		c.OnDetach(entity)
		if c.Base().Attachments == 0 {
			c.OnDelete()
			db.columns[id].Detach(c.Base().indexInColumn)
		}
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

func (db *ECS) DeserializeAndAttachEntity(jsonEntityComponents map[string]any) {
	var jsonEntity string
	var entity Entity
	var ok bool
	var err error
	if jsonEntityComponents["Entity"] == nil {
		log.Printf("ECS.DeserializeAndAttachEntity: json object doesn't have entity key")
		return
	}
	if jsonEntity, ok = jsonEntityComponents["Entity"].(string); !ok {
		log.Printf("ECS.DeserializeAndAttachEntity: json entity isn't string")
		return
	}
	if entity, err = ParseEntity(jsonEntity); err != nil {
		log.Printf("ECS.DeserializeAndAttachEntity: json entity can't be parsed: %v", err)
		return
	}

	db.Entities.Set(uint32(entity))

	for name, cid := range Types().IDs {
		jsonData := jsonEntityComponents[name]
		if jsonData == nil {
			continue
		}
		if jsonLink, ok := jsonData.(string); ok {
			linkedEntity, _ := ParseEntity(jsonLink)
			if linkedEntity != 0 {
				c := db.Component(linkedEntity, cid)
				if c != nil {
					db.attach(entity, c, cid)
				}
			}
		} else {
			jsonComponent := jsonData.(map[string]any)
			attached := db.attach(entity, nil, cid)
			if attached.Base().Attachments == 1 {
				attached.Construct(jsonComponent)
			}
		}
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
	err = yaml.Unmarshal(fileContents, &parsed)
	//err = json.Unmarshal(fileContents, &parsed)
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

func (db *ECS) serializeEntity(entity Entity, savedComponents map[uint64]Entity) map[string]any {
	jsonEntity := make(map[string]any)
	jsonEntity["Entity"] = entity.String()
	for _, component := range db.rows[int(entity)] {
		if component == nil || component.IsSystem() {
			continue
		}
		cid := component.Base().ComponentID
		hash := (uint64(component.Base().indexInColumn) << 16) | (uint64(cid) & 0xFFFF)
		col := Types().ColumnPlaceholders[cid]
		jsonID := col.Type().String()

		if savedComponents == nil {
			jsonComponent := component.Serialize()
			delete(jsonComponent, "Entities")
			jsonEntity[jsonID] = jsonComponent
		} else if savedEntity, ok := savedComponents[hash]; ok {
			jsonEntity[jsonID] = savedEntity.String()
		} else {
			jsonComponent := component.Serialize()
			delete(jsonComponent, "Entities")
			jsonEntity[jsonID] = jsonComponent
			savedComponents[hash] = entity
		}

	}
	if len(jsonEntity) == 1 {
		return nil
	}
	return jsonEntity
}

func (db *ECS) SerializeEntity(entity Entity) map[string]any {
	return db.serializeEntity(entity, nil)
}

func (db *ECS) Save(filename string) {
	db.Lock.Lock()
	defer db.Lock.Unlock()
	jsonECS := make([]any, 0)
	savedComponents := make(map[uint64]Entity)

	db.Entities.Range(func(entity uint32) {
		jsonEntity := db.serializeEntity(Entity(entity), savedComponents)
		if len(jsonEntity) == 0 {
			return
		}
		jsonECS = append(jsonECS, jsonEntity)
	})

	bytes, err := yaml.Marshal(jsonECS)
	//bytes, err := json.MarshalIndent(jsonECS, "", "  ")

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
