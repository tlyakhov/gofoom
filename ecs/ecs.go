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
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/dynamic"

	"github.com/kelindar/bitmap"
	"sigs.k8s.io/yaml"
)

// The architecture is like this:
// * An entity is a globally unique integer (uint32), e.g. primary key
// * An entity can be associated with multiple components (one of each kind)
// * A system (controller) is code that queries and operates on components and
// entities
var (
	Simulation      *dynamic.Simulation
	Entities        bitmap.Bitmap
	Lock            sync.RWMutex
	FuncMap         template.FuncMap
	SourceFileNames map[string]*SourceFile
	SourceFileIDs   map[EntitySourceID]*SourceFile

	rows   [EntitySourceIDBits][]ComponentTable
	arenas []ComponentArena
)

func Initialize() {
	defer concepts.ExecutionDuration(concepts.ExecutionTrack("ecs.Initialize"))
	Lock.Lock()
	defer Lock.Unlock()
	// We may have existing entities and components. Let's run any delete
	// functions to clean up stuff like non-Go data.
	for i, arena := range arenas {
		if i == 0 {
			continue
		}
		for j := range arena.Cap() {
			a := arena.Component(j)
			if a == nil {
				continue
			}
			for _, e := range a.Base().Entities {
				if e != 0 {
					a.OnDetach(e)
				}
			}
			arena.Detach(j)
			a.OnDelete()
		}
	}
	for i := range len(rows) {
		rows[i] = nil
	}
	Entities = bitmap.Bitmap{}
	// 0 is reserved and represents 'null' entity
	Entities.Set(0)
	// 1 is reserved and represents the root file.
	CreateEntity(1)
	arenas = make([]ComponentArena, len(Types().ArenaPlaceholders))

	// Preserve simulation callbacks
	sim := dynamic.NewSimulation()
	if Simulation != nil {
		sim.Integrate = Simulation.Integrate
		sim.Render = Simulation.Render
		sim.NewFrame = Simulation.NewFrame
	}
	Simulation = sim

	SourceFileNames = make(map[string]*SourceFile)
	SourceFileIDs = make(map[EntitySourceID]*SourceFile)
	FuncMap = template.FuncMap{}

	// Initialize component arenas based on registered component types.
	for i, arenaPlaceholder := range Types().ArenaPlaceholders {
		if arenaPlaceholder == nil {
			continue
		}
		// log.Printf("Component %v, index: %v", arenaPlaceholder.Type().String(), i)
		// t = *ComponentArena[T]
		t := reflect.TypeOf(arenaPlaceholder)
		arenas[i] = reflect.New(t.Elem()).Interface().(ComponentArena)
		arenas[i].From(arenaPlaceholder)
		fmName := arenaPlaceholder.Type().String()
		fmName = strings.ReplaceAll(fmName, ".", "_")
		FuncMap[fmName] = func(e Entity) Component { return GetComponent(e, arenaPlaceholder.ID()) }
	}

	/*for _, meta := range Types().Controllers {
		fmt.Printf("%v, priority %v\n", meta.Type.String(), meta.Priority)
	}*/
}

// Reserves an entity ID in the database (no components attached)
// It finds the smallest available entity ID, marks it as used, and returns it.
func NewEntity() Entity {
	var nextFree uint32
	var found bool
	if nextFree, found = Entities.MinZero(); !found {
		nextFree = uint32(len(rows[0]))
	}
	Entities.Set(nextFree)
	expandRows(0, int(nextFree+1))
	return Entity(nextFree)
}

// NextFreeEntitySourceID returns the next available entity source ID.
// It iterates through all possible source IDs and returns the first one that
// is not currently in use.
func NextFreeEntitySourceID() EntitySourceID {
	for i := range 1 << EntitySourceIDBits {
		id := EntitySourceID(i)
		if _, ok := SourceFileIDs[id]; !ok {
			return id
		}
	}
	return EntitySourceID(1<<EntitySourceIDBits - 1)
}

func ArenaFor[T any, PT GenericAttachable[T]](id ComponentID) *Arena[T, PT] {
	return arenas[id].(*Arena[T, PT])
}

func ArenaByID(id ComponentID) ComponentArena {
	return arenas[id]
}

func localizeEntity(entity Entity) (sid EntitySourceID, local Entity) {
	if entity == 0 {
		return 0, 0
	}
	sid, local = entity.SourceID(), entity.Local()
	if len(rows[sid]) <= int(local) {
		return 0, 0
	}
	return
}

// AllComponents retrieves the component table for a specific entity.
func AllComponents(entity Entity) ComponentTable {
	if sid, local := localizeEntity(entity); local != 0 {
		return rows[sid][int(local)]
	}
	return nil
}

// Callers need to be careful, this function can return nil that's not castable
// to an actual component type. The Get* methods are better.
func GetComponent(entity Entity, id ComponentID) Component {
	if id == 0 {
		return nil
	}
	if sid, local := localizeEntity(entity); local != 0 {
		return rows[sid][int(local)].Get(id)
	}
	return nil
}

func Singleton(id ComponentID) Component {
	c := arenas[id]
	if c.Len() != 0 {
		return c.Component(0)
	}
	return NewAttachedComponent(NewEntity(), id)
}

func First(id ComponentID) Component {
	c := arenas[id]
	for i := range c.Cap() {
		a := arenas[id].Component(i)
		if a != nil {
			return a
		}
	}
	return nil
}

func Link(target Entity, source Entity) {
	sourceID, sourceLocal := localizeEntity(source)
	if sourceLocal == 0 {
		return
	}
	_, targetLocal := localizeEntity(target)
	if targetLocal == 0 {
		return
	}
	for _, c := range rows[sourceID][int(sourceLocal)] {
		if c == nil || !c.MultiAttachable() {
			continue
		}
		attach(target, &c, c.ComponentID())
	}
}

// TODO: This approach of reallocating and copying is probably not scalable.
// We should benchmark this for large entity counts and potentially replace with
// some kind of chunked arena, or at least tweak the grow rate.
func expandRows(sid EntitySourceID, size int) {
	if size < cap(rows[sid]) {
		for size >= len(rows[sid]) {
			rows[sid] = append(rows[sid], nil)
		}
	} else {
		newSize := (((size + 1) / 256) + 1) * 256
		tmp := rows[sid]
		rows[sid] = make([]ComponentTable, size+1, newSize)
		copy(rows[sid], tmp)
	}
}

func CreateEntity(entity Entity) {
	sid, local := entity.SourceID(), entity.Local()
	expandRows(sid, int(local))
	Entities.Set(uint32(entity))
}

// Attach a component to an entity. If a component with this type is already
// attached, this method will overwrite it.
func attach(entity Entity, component *Component, componentID ComponentID) {
	if entity == 0 {
		log.Printf("ecs.attach: tried to attach 0 entity!")
		return
	}

	if componentID == 0 {
		log.Printf("ecs.attach: tried to attach 0 component ID!")
		return
	}

	sid, local := entity.SourceID(), entity.Local()
	expandRows(sid, int(local))

	if !Entities.Contains(uint32(entity)) {
		log.Printf("ecs.attach: warning: ecs.Entities does not contain entity.")
	}

	// Try to retrieve the existing component for this entity
	ec := rows[sid][int(local)].Get(componentID)

	// Did the caller:
	// 1. not provide a component?
	// 2. the provided component is unattached?
	if *component == nil || (*component).Base().Attachments == 0 {
		// Then we need to add a new element to the arena:
		arena := arenas[componentID]
		if ec != nil {
			// A component with this index is already attached to this entity, overwrite it.
			indexInArena := ec.Base().indexInArena
			arena.Replace(component, indexInArena)
		} else {
			// This entity doesn't have a component with this index attached. Extend the
			// slice.
			arena.Add(component)
		}
	} else if ec != nil {
		// We have a conflict between the provided component and an existing one
		// with the same component ID. We should abort. This happens with Linked components.
		// log.Printf("ecs.attach: Entity %v already has a component %v. Aborting!", entity, Types().ArenaPlaceholders[componentID].String())
		return
	}

	attachable := *component
	a := attachable.Base()
	if a.Attachments > 0 && !attachable.MultiAttachable() {
		log.Printf("ecs.attach: tried to attach Component %v to %v, but it is already attached to %v and not multi-attachable.", attachable.String(), entity, a.Entity.ShortString())
		return
	}
	a.Entities.Set(entity)
	a.Entity = entity
	a.Attachments++
	rows[sid][int(local)].Set(attachable)
	attachable.OnAttach()
}

// Create a new component with the given index and attach it.
func NewAttachedComponent(entity Entity, id ComponentID) Component {
	var attached Component
	attach(entity, &attached, id)
	attached.Construct(nil)
	return attached
}

func LoadComponentWithoutAttaching(id ComponentID, data map[string]any) Component {
	if data == nil {
		return nil
	}
	component := Types().ArenaPlaceholders[id].New()
	component.Construct(data)
	return component
}

func NewAttachedComponentTyped(entity Entity, cType string) Component {
	if index, ok := Types().IDs[cType]; ok {
		return NewAttachedComponent(entity, index)
	}

	log.Printf("NewComponent: unregistered type %v for entity %v\n", cType, entity)
	return nil
}

// Attach a component to an entity. `component` is a pointer to an interface
// because it's both an input and output - you can provide an entire component
// to attach or a pointer to nil to get back a new one. Previously this method
// had semantics like Go's `append`, but this was too error prone if the return
// value was ignored.
func Attach(id ComponentID, entity Entity, component *Component) {
	attach(entity, component, id)
}

func AttachTyped[T any, PT GenericAttachable[T]](entity Entity, component *PT) {
	var attachable Component
	if *component != nil {
		attachable = *component
		attach(entity, &attachable, attachable.ComponentID())
	} else {
		cid := Types().IDs[reflect.TypeFor[PT]().String()]
		attach(entity, &attachable, cid)
	}
	*component = attachable.(PT)
}

func detach(id ComponentID, entity Entity, checkForEmpty bool) {
	if entity == 0 {
		log.Printf("ecs.Detach: tried to detach 0 entity.")
		return
	}
	if id == 0 {
		log.Printf("ecs.Detach: tried to detach 0 component index.")
		return
	}

	sid, local := entity.SourceID(), entity.Local()

	if len(rows[sid]) <= int(local) {
		log.Printf("ecs.Detach: entity %v is >= length of list %v.", local, len(rows[sid]))
		return
	}
	ec := rows[sid][int(local)].Get(id)
	arena := arenas[id]
	if ec == nil {
		// This component is not attached
		log.Printf("ecs.Detach: tried to detach unattached component %v from entity %v", arena.String(), entity)
		return
	}

	ec.OnDetach(entity)
	if ec.Base().Attachments == 0 {
		arena.Detach(ec.Base().indexInArena)
		ec.OnDelete()
	}
	rows[sid][int(local)].Delete(id)

	if checkForEmpty {
		allNil := true
		for _, a := range rows[sid][int(local)] {
			if a != nil {
				allNil = false
				break
			}
		}
		if allNil {
			Entities.Remove(uint32(entity))
			rows[sid][int(local)] = nil
		}
	}
}

func DetachComponent(id ComponentID, entity Entity) {
	detach(id, entity, true)
}

func DeleteByType(component Component) {
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
			detach(id, entity, true)
		}
	}
}

func Delete(entity Entity) {
	if entity == 0 {
		return
	}

	Entities.Remove(uint32(entity))

	sid, local := localizeEntity(entity)
	if local == 0 {
		return
	}

	for _, c := range rows[sid][int(local)] {
		if c == nil {
			continue
		}
		id := c.ComponentID()
		c.OnDetach(entity)
		if c.Base().Attachments == 0 {
			c.OnDelete()
			arenas[id].Detach(c.Base().indexInArena)
		}
	}
	rows[sid][int(local)] = nil
}

// TODO: Optimize this and add ability to wildcard search
func GetEntityByName(name string) Entity {
	arena := ArenaFor[Named](NamedCID)
	for i := range arena.Cap() {
		if named := arena.Value(i); named != nil && named.Name == name {
			return named.Entity
		}
	}

	return 0
}

func Load(filename string) error {
	// Root file is attached to entity 1
	file := NewAttachedComponent(1, SourceFileCID).(*SourceFile)
	file.Source = filename
	file.ID = 0
	file.Flags = EntityInternal
	return file.Load()
}

func Save(filename string) {
	bytes, err := yaml.Marshal(SaveSnapshot(false))
	//bytes, err := json.MarshalIndent(yamlECS, "", "  ")

	if err != nil {
		panic(err)
	}

	os.WriteFile(filename, bytes, os.ModePerm)
}

// Returns true if a new one was created.
func CachedGeneratedComponent[T any, PT GenericAttachable[T]](field *PT, name string, cid ComponentID) bool {
	if *field != nil {
		return false
	}
	e := GetEntityByName(name)
	if e != 0 {
		*field = GetComponent(e, cid).(PT)
		if *field != nil {
			return false
		}
	} else {
		e = NewEntity()
	}

	*field = NewAttachedComponent(e, cid).(PT)
	base := (*field).Base()
	base.Flags |= EntityInternal
	n := NewAttachedComponent(base.Entity, NamedCID).(*Named)
	n.Name = name
	n.Flags |= EntityInternal

	return true
}
