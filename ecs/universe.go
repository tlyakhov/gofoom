// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
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
type Universe struct {
	*dynamic.Simulation
	Entities        bitmap.Bitmap
	Lock            sync.RWMutex
	FuncMap         template.FuncMap
	SourceFileNames map[string]*SourceFile
	SourceFileIDs   map[EntitySourceID]*SourceFile
	rows            []ComponentTable
	columns         []AttachableColumn
}

func NewUniverse() *Universe {
	u := &Universe{}
	u.Clear()

	return u
}

func (u *Universe) Clear() {
	u.Entities = bitmap.Bitmap{}
	// 0 is reserved and represents 'null' entity
	u.Entities.Set(0)
	// rows are indexed by entity ID, so we need to reserve the 0th row
	u.rows = make([]ComponentTable, 1)
	u.columns = make([]AttachableColumn, len(Types().ColumnPlaceholders))
	u.Simulation = dynamic.NewSimulation()
	u.SourceFileNames = make(map[string]*SourceFile)
	u.SourceFileIDs = make(map[EntitySourceID]*SourceFile)
	u.FuncMap = template.FuncMap{
		"Universe": func() *Universe { return u },
	}

	// Initialize component columns based on registered component types.
	for i, columnPlaceholder := range Types().ColumnPlaceholders {
		if columnPlaceholder == nil {
			continue
		}
		// log.Printf("Component %v, index: %v", columnPlaceholder.Type().String(), i)
		// t = *ComponentColumn[T]
		t := reflect.TypeOf(columnPlaceholder)
		u.columns[i] = reflect.New(t.Elem()).Interface().(AttachableColumn)
		u.columns[i].From(columnPlaceholder, u)
		fmName := columnPlaceholder.Type().String()
		fmName = strings.ReplaceAll(fmName, ".", "_")
		u.FuncMap[fmName] = func(e Entity) Attachable { return u.Component(e, columnPlaceholder.ID()) }
	}
}

// Reserves an entity ID in the database (no components attached)
// It finds the smallest available entity ID, marks it as used, and returns it.
func (u *Universe) NewEntity() Entity {
	if free, found := u.Entities.MinZero(); found {
		u.Entities.Set(free)
		return Entity(free)
	}
	nextFree := len(u.rows)
	for len(u.rows) < (nextFree + 1) {
		u.rows = append(u.rows, nil)
	}
	u.Entities.Set(uint32(nextFree))
	return Entity(nextFree)
}

// NextFreeEntitySourceID returns the next available entity source ID.
// It iterates through all possible source IDs and returns the first one that
// is not currently in use.
func (u *Universe) NextFreeEntitySourceID() EntitySourceID {
	for i := range 1 << EntitySourceIDBits {
		id := EntitySourceID(i)
		if _, ok := u.SourceFileIDs[id]; !ok {
			return id
		}
	}
	return EntitySourceID(1<<EntitySourceIDBits - 1)
}

func ColumnFor[T any, PT GenericAttachable[T]](u *Universe, id ComponentID) *Column[T, PT] {
	return u.columns[id].(*Column[T, PT])
}

func (u *Universe) Column(id ComponentID) AttachableColumn {
	return u.columns[id]
}

// AllComponents retrieves the component table for a specific entity.
func (u *Universe) AllComponents(entity Entity) ComponentTable {
	if entity == 0 || len(u.rows) <= int(entity) {
		return nil
	}
	return u.rows[int(entity)]
}

// Callers need to be careful, this function can return nil that's not castable
// to an actual component type. The Get* methods are better.
func (u *Universe) Component(entity Entity, id ComponentID) Attachable {
	if entity == 0 || id == 0 || u == nil || len(u.rows) <= int(entity) {
		return nil
	}
	return u.rows[int(entity)].Get(id)
}

func (u *Universe) Singleton(id ComponentID) Attachable {
	c := u.columns[id]
	if c.Len() != 0 {
		return c.Attachable(0)
	}
	return u.NewAttachedComponent(u.NewEntity(), id)
}

func (u *Universe) First(id ComponentID) Attachable {
	c := u.columns[id]
	for i := range c.Cap() {
		a := u.columns[id].Attachable(i)
		if a != nil {
			return a
		}
	}
	return nil
}

func (u *Universe) Link(target Entity, source Entity) {
	if target == 0 || source == 0 ||
		len(u.rows) <= int(source) || len(u.rows) <= int(target) {
		return
	}
	for _, c := range u.rows[int(source)] {
		if c == nil || !c.MultiAttachable() {
			continue
		}
		u.attach(target, &c, c.Base().ComponentID)
	}
}

// Attach a component to an entity. If a component with this type is already
// attached, this method will overwrite it.
func (u *Universe) attach(entity Entity, component *Attachable, componentID ComponentID) {
	if entity == 0 {
		log.Printf("Universe.attach: tried to attach 0 entity!")
		return
	}

	if componentID == 0 {
		log.Printf("Universe.attach: tried to attach 0 component ID!")
		return
	}

	for int(entity) >= len(u.rows) {
		u.rows = append(u.rows, nil)
	}

	// Try to retrieve the existing component for this entity
	ec := u.rows[int(entity)].Get(componentID)

	// Did the caller:
	// 1. not provide a component?
	// 2. the provided component is unattached?
	if *component == nil || (*component).Base().Attachments == 0 {
		// Then we need to add a new element to the column:
		column := u.columns[componentID]
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
		// log.Printf("Universe.attach: Entity %v already has a component %v. Aborting!", entity, Types().ColumnPlaceholders[componentID].String())
		return
	}

	attachable := *component
	a := attachable.Base()
	if a.Attachments > 0 && !attachable.MultiAttachable() {
		log.Printf("Universe.attach: Component %v is already attached to %v and not multi-attachable.", attachable.String(), a.Entity.ShortString())
	}
	a.Entities.Set(entity)
	a.Entity = entity
	a.Attachments++
	a.ComponentID = componentID
	u.rows[int(entity)].Set(attachable)
	attachable.OnAttach(u)
}

// Create a new component with the given index and attach it.
func (u *Universe) NewAttachedComponent(entity Entity, id ComponentID) Attachable {
	var attached Attachable
	u.attach(entity, &attached, id)
	attached.Construct(nil)
	return attached
}

func (u *Universe) LoadComponentWithoutAttaching(id ComponentID, data map[string]any) Attachable {
	if data == nil {
		return nil
	}
	component := Types().ColumnPlaceholders[id].New()
	component.Base().Universe = u
	component.Construct(data)
	return component
}

func (u *Universe) NewAttachedComponentTyped(entity Entity, cType string) Attachable {
	if index, ok := Types().IDs[cType]; ok {
		return u.NewAttachedComponent(entity, index)
	}

	log.Printf("NewComponent: unregistered type %v for entity %v\n", cType, entity)
	return nil
}

// Attach a component to an entity. `component` is a pointer to an interface
// because it's both an input and output - you can provide an entire component
// to attach or a pointer to nil to get back a new one. Previously this method
// had semantics like Go's `append`, but this was too error prone if the return
// value was ignored.
func (u *Universe) Attach(id ComponentID, entity Entity, component *Attachable) {
	u.attach(entity, component, id)
}

func AttachTyped[T any, PT GenericAttachable[T]](u *Universe, entity Entity, component *PT) {
	var attachable Attachable
	if *component != nil {
		attachable = *component
		u.attach(entity, &attachable, attachable.Base().ComponentID)
	} else {
		cid := Types().IDs[reflect.TypeFor[PT]().String()]
		u.attach(entity, &attachable, cid)
	}
	*component = attachable.(PT)
}

func (u *Universe) detach(id ComponentID, entity Entity, checkForEmpty bool) {
	if entity == 0 {
		log.Printf("Universe.Detach: tried to detach 0 entity.")
		return
	}
	if id == 0 {
		log.Printf("Universe.Detach: tried to detach 0 component index.")
		return
	}

	if len(u.rows) <= int(entity) {
		log.Printf("Universe.Detach: entity %v is >= length of list %v.", entity, len(u.rows))
		return
	}
	ec := u.rows[int(entity)].Get(id)
	column := u.columns[id]
	if ec == nil {
		// This component is not attached
		log.Printf("Universe.Detach: tried to detach unattached component %v from entity %v", column.String(), entity)
		return
	}

	ec.OnDetach(entity)
	if ec.Base().Attachments == 0 {
		column.Detach(ec.Base().indexInColumn)
		ec.OnDelete()
	}
	u.rows[int(entity)].Delete(id)

	if checkForEmpty {
		allNil := true
		for _, a := range u.rows[int(entity)] {
			if a != nil {
				allNil = false
				break
			}
		}
		if allNil {
			u.Entities.Remove(uint32(entity))
			u.rows[int(entity)] = nil
		}
	}
}

func (u *Universe) DetachComponent(id ComponentID, entity Entity) {
	u.detach(id, entity, true)
}

func (u *Universe) DeleteByType(component Attachable) {
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
			u.detach(id, entity, true)
		}
	}
}

func (u *Universe) Delete(entity Entity) {
	if entity == 0 {
		return
	}

	u.Entities.Remove(uint32(entity))

	if len(u.rows) <= int(entity) {
		return
	}

	for _, c := range u.rows[int(entity)] {
		if c == nil {
			continue
		}
		id := c.Base().ComponentID
		c.OnDetach(entity)
		if c.Base().Attachments == 0 {
			c.OnDelete()
			u.columns[id].Detach(c.Base().indexInColumn)
		}
	}
	u.rows[int(entity)] = nil
}

// TODO: Optimize this and add ability to wildcard search
func (u *Universe) GetEntityByName(name string) Entity {
	col := ColumnFor[Named](u, NamedCID)
	for i := range col.Cap() {
		if named := col.Value(i); named != nil && named.Name == name {
			return named.Entity
		}
	}

	return 0
}

func (u *Universe) EntityAllNoSave(entity Entity) bool {
	if entity == 0 || len(u.rows) <= int(entity) {
		return false
	}
	for _, c := range u.rows[int(entity)] {
		if c == nil {
			continue
		}
		if c.Base().Flags&ComponentNoSave == 0 {
			return false
		}
	}
	return true
}

func (u *Universe) Load(filename string) error {
	file := u.NewAttachedComponent(u.NewEntity(), SourceFileCID).(*SourceFile)
	file.Source = filename
	file.ID = 0
	file.Flags = ComponentInternal
	return file.Load()
}

func (u *Universe) serializeEntity(entity Entity, savedComponents map[uint64]Entity) map[string]any {
	yamlEntity := make(map[string]any)
	yamlEntity["Entity"] = entity.String()
	for _, component := range u.rows[int(entity)] {
		if component == nil || (component.Base().Flags&ComponentNoSave != 0) {
			continue
		}
		cid := component.Base().ComponentID
		hash := (uint64(component.Base().indexInColumn) << 16) | (uint64(cid) & 0xFFFF)
		col := Types().ColumnPlaceholders[cid]
		yamlID := col.Type().String()

		if savedComponents != nil {
			if savedEntity, ok := savedComponents[hash]; ok {
				yamlEntity[yamlID] = savedEntity.Serialize(u)
				continue
			}
		}

		if component.Base().IsExternal() {
			// Just pick one
			// TODO: This has a code smell. Should there be a particular way
			// to pick an entity ID to reference when saving?
			yamlEntity[yamlID] = component.Base().ExternalEntities()[0].Serialize(u)
			continue
		}

		yamlComponent := component.Serialize()
		delete(yamlComponent, "Entities")
		yamlEntity[yamlID] = yamlComponent
		if savedComponents != nil {
			savedComponents[hash] = entity
		}

	}
	if len(yamlEntity) == 1 {
		return nil
	}
	return yamlEntity
}

func (u *Universe) SerializeEntity(entity Entity) map[string]any {
	return u.serializeEntity(entity, nil)
}

func (u *Universe) Save(filename string) {
	u.Lock.Lock()
	defer u.Lock.Unlock()
	yamlECS := make([]any, 0)
	savedComponents := make(map[uint64]Entity)

	u.Entities.Range(func(entity uint32) {
		e := Entity(entity)
		if e.IsExternal() {
			return
		}
		yamlEntity := u.serializeEntity(e, savedComponents)
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

// Returns true if a new one was created.
func CachedGeneratedComponent[T any, PT GenericAttachable[T]](u *Universe, field *PT, name string, cid ComponentID) bool {
	if *field != nil {
		return false
	}
	e := u.GetEntityByName(name)
	if e != 0 {
		*field = u.Component(e, cid).(PT)
		if *field != nil {
			return false
		}
	} else {
		e = u.NewEntity()
	}

	*field = u.NewAttachedComponent(e, cid).(PT)
	base := (*field).Base()
	base.Flags = ComponentInternal
	n := u.NewAttachedComponent(base.Entity, NamedCID).(*Named)
	n.Name = name
	n.Flags = ComponentInternal

	return true
}
