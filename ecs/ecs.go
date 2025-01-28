// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"reflect"
	"strings"
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
	FuncMap  template.FuncMap

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
	db.FuncMap = template.FuncMap{
		"ECS": func() *ECS { return db },
	}

	for i, columnPlaceholder := range Types().ColumnPlaceholders {
		if columnPlaceholder == nil {
			continue
		}
		// log.Printf("Component %v, index: %v", columnPlaceholder.Type().String(), i)
		// t = *ComponentColumn[T]
		t := reflect.TypeOf(columnPlaceholder)
		db.columns[i] = reflect.New(t.Elem()).Interface().(AttachableColumn)
		db.columns[i].From(columnPlaceholder, db)
		fmName := columnPlaceholder.Type().String()
		fmName = strings.ReplaceAll(fmName, ".", "_")
		db.FuncMap[fmName] = func(e Entity) Attachable { return db.Component(e, columnPlaceholder.ID()) }
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

func (db *ECS) Singleton(id ComponentID) Attachable {
	c := db.columns[id]
	if c.Len() != 0 {
		return c.Attachable(0)
	}
	return db.NewAttachedComponent(db.NewEntity(), id)
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
		db.attach(target, &c, c.Base().ComponentID)
	}
}

// Attach a component to an entity. If a component with this type is already
// attached, this method will overwrite it.
func (db *ECS) attach(entity Entity, component *Attachable, componentID ComponentID) {
	if entity == 0 {
		log.Printf("ECS.attach: tried to attach 0 entity!")
		return
	}

	if componentID == 0 {
		log.Printf("ECS.attach: tried to attach 0 component ID!")
		return
	}

	for int(entity) >= len(db.rows) {
		db.rows = append(db.rows, nil)
	}

	// Try to retrieve the existing component for this entity
	ec := db.rows[int(entity)].Get(componentID)

	// Did the caller:
	// 1. not provide a component?
	// 2. the provided component is unattached?
	if *component == nil || (*component).Base().Attachments == 0 {
		// Then we need to add a new element to the column:
		column := db.columns[componentID]
		if ec != nil {
			// A component with this index is already attached to this entity, overwrite it.
			indexInColumn := ec.Base().indexInColumn
			column.Replace(component, indexInColumn)
		} else {
			// This entity doesn't have a component with this index attached. Extend the
			// slice.
			column.Add(component)
		}
	} else if ec != nil {
		// We have a conflict between the provided component and an existing one
		// with the same component ID. We should abort. This happens with Linked components.
		// log.Printf("ECS.attach: Entity %v already has a component %v. Aborting!", entity, Types().ColumnPlaceholders[componentID].String())
		return
	}

	attachable := *component
	a := attachable.Base()
	if a.Attachments > 0 && !attachable.MultiAttachable() {
		log.Printf("ECS.attach: Component %v is already attached to %v and not multi-attachable.", attachable.String(), a.Entity)
	}
	a.Entities.Set(entity)
	a.Entity = entity
	a.Attachments++
	a.ComponentID = componentID
	db.rows[int(entity)].Set(attachable)
	attachable.OnAttach(db)
}

// Create a new component with the given index and attach it.
func (db *ECS) NewAttachedComponent(entity Entity, id ComponentID) Attachable {
	var attached Attachable
	db.attach(entity, &attached, id)
	attached.Construct(nil)
	return attached
}

func (db *ECS) LoadAttachComponent(id ComponentID, data map[string]any, ignoreSerializedEntity bool) Attachable {
	entities, attachments := ParseEntitiesFromMap(data)
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
		db.attach(entity, &attached, id)
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
	component.OnAttach(db)
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

// Attach a component to an entity. `component` is a pointer to an interface
// because it's both an input and output - you can provide an entire component
// to attach or a pointer to nil to get back a new one. Previously this method
// had semantics like Go's `append`, but this was too error prone if the return
// value was ignored.
func (db *ECS) Attach(id ComponentID, entity Entity, component *Attachable) {
	db.attach(entity, component, id)
}

func AttachTyped[T any, PT GenericAttachable[T]](db *ECS, entity Entity, component *PT) {
	attachable := Attachable(*component)
	db.attach(entity, &attachable, attachable.Base().ComponentID)
	*component = attachable.(PT)
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

func (db *ECS) EntityAllSystem(entity Entity) bool {
	if entity == 0 || len(db.rows) <= int(entity) {
		return false
	}
	for _, c := range db.rows[int(entity)] {
		if c == nil {
			continue
		}
		if !c.IsSystem() {
			return false
		}
	}
	return true
}

func (db *ECS) DeserializeAndAttachEntity(yamlEntityComponents map[string]any) {
	var yamlEntity string
	var entity Entity
	var ok bool
	var err error
	if yamlEntityComponents["Entity"] == nil {
		log.Printf("ECS.DeserializeAndAttachEntity: yaml object doesn't have entity key")
		return
	}
	if yamlEntity, ok = yamlEntityComponents["Entity"].(string); !ok {
		log.Printf("ECS.DeserializeAndAttachEntity: yaml entity isn't string")
		return
	}
	if entity, err = ParseEntity(yamlEntity); err != nil {
		log.Printf("ECS.DeserializeAndAttachEntity: yaml entity can't be parsed: %v", err)
		return
	}

	db.Entities.Set(uint32(entity))

	for name, cid := range Types().IDs {
		yamlData := yamlEntityComponents[name]
		if yamlData == nil {
			continue
		}
		if yamlLink, ok := yamlData.(string); ok {
			linkedEntity, _ := ParseEntity(yamlLink)
			if linkedEntity != 0 {
				c := db.Component(linkedEntity, cid)
				if c != nil {
					db.attach(entity, &c, cid)
				}
			}
		} else {
			yamlComponent := yamlData.(map[string]any)
			var attached Attachable
			db.attach(entity, &attached, cid)
			if attached.Base().Attachments == 1 {
				attached.Construct(yamlComponent)
			}
		}
	}
}

/***

	TODO: Something really useful could be some kind of #include directive to be
	able to combine multiple files into one structure. This way, common game
	elements (say, all the various monster types, items, etc...) could be set up
	in a "prefab" world file, and then referenced from actual levels/missions.

	This has a lot of potential complexity - for example, how do we map entity
	IDs across file boundaries? And then, after mapping, what happens when the
	new combined world is saved or edited? How do we ensure that the "imported"
	data stays untouched?

	Seems like we would need a few things for this:
	1. When loading, walk the YAML looking for entity IDs, creating a map of
	   original->loaded IDs.
	2. Some kind of "locked/included" flag at the entity level to prevent
	   operations that would modify the original data (e.g. attaching new components)
	3. When linking non-included and included entities together (via components
	   or references), we would need both the mapped and original IDs there, since
	   the mapped IDs are dynamic. This is kind of the worst part of it.

**/

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

	var yamlEntities []any
	var ok bool
	if yamlEntities, ok = parsed.([]any); !ok || yamlEntities == nil {
		return fmt.Errorf("ECS.Load: YAML root must be an array")
	}

	for _, yamlData := range yamlEntities {
		yamlEntity := yamlData.(map[string]any)
		if yamlEntity == nil {
			log.Printf("ECS.Load: YAML array element should be an object")
			continue
		}
		db.DeserializeAndAttachEntity(yamlEntity)
	}

	// After everything's loaded, trigger the controllers
	db.ActAllControllers(ControllerRecalculate)
	return nil
}

func (db *ECS) serializeEntity(entity Entity, savedComponents map[uint64]Entity) map[string]any {
	yamlEntity := make(map[string]any)
	yamlEntity["Entity"] = entity.String()
	for _, component := range db.rows[int(entity)] {
		if component == nil || component.IsSystem() {
			continue
		}
		cid := component.Base().ComponentID
		hash := (uint64(component.Base().indexInColumn) << 16) | (uint64(cid) & 0xFFFF)
		col := Types().ColumnPlaceholders[cid]
		yamlID := col.Type().String()

		if savedComponents == nil {
			yamlComponent := component.Serialize()
			delete(yamlComponent, "Entities")
			yamlEntity[yamlID] = yamlComponent
		} else if savedEntity, ok := savedComponents[hash]; ok {
			yamlEntity[yamlID] = savedEntity.String()
		} else {
			yamlComponent := component.Serialize()
			delete(yamlComponent, "Entities")
			yamlEntity[yamlID] = yamlComponent
			savedComponents[hash] = entity
		}

	}
	if len(yamlEntity) == 1 {
		return nil
	}
	return yamlEntity
}

func (db *ECS) SerializeEntity(entity Entity) map[string]any {
	return db.serializeEntity(entity, nil)
}

func (db *ECS) Save(filename string) {
	db.Lock.Lock()
	defer db.Lock.Unlock()
	yamlECS := make([]any, 0)
	savedComponents := make(map[uint64]Entity)

	db.Entities.Range(func(entity uint32) {
		log.Printf("e: %v", entity)
		yamlEntity := db.serializeEntity(Entity(entity), savedComponents)
		if len(yamlEntity) == 0 {
			return
		}
		yamlECS = append(yamlECS, yamlEntity)
	})

	bytes, err := yaml.Marshal(yamlECS)
	//bytes, err := json.MarshalIndent(yamlECS, "", "  ")

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
