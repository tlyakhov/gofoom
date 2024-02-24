package concepts

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"slices"
	"strconv"
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
	Components       [][]Attachable
	EntityComponents [][]Attachable
	usedEntities     bitmap.Bitmap
	Lock             sync.RWMutex
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
	db.Components = make([][]Attachable, len(DbTypes().Types))
	db.Simulation = NewSimulation()
	for i := 0; i < len(DbTypes().Types); i++ {
		db.Components[i] = make([]Attachable, 0)
	}
}

// Reserves an entity ID in the database (no components attached)
func (db *EntityComponentDB) NewEntity() uint64 {
	if free, found := db.usedEntities.MinZero(); found {
		db.usedEntities.Set(free)
		return uint64(free)
	}
	nextFree := len(db.EntityComponents)
	db.EntityComponents = append(db.EntityComponents, nil)
	db.usedEntities.Set(uint32(nextFree))
	return uint64(nextFree)
}

// Reserves an entity ID in the database and returns a reference to it
func (db *EntityComponentDB) RefForNewEntity() *EntityRef {
	entity := db.NewEntity()
	return &EntityRef{Entity: entity, DB: db}
}

func (db *EntityComponentDB) EntityRef(entity uint64) *EntityRef {
	return &EntityRef{DB: db, Entity: entity}
}

func (db *EntityComponentDB) All(index int) []Attachable {
	return db.Components[index]
}

func (db *EntityComponentDB) AllForType(cType string) []Attachable {
	if index, ok := DbTypes().Indexes[cType]; ok {
		return db.Components[index]
	}
	return nil
}

func (db *EntityComponentDB) First(index int) Attachable {
	for _, c := range db.Components[index] {
		return c
	}
	return nil
}

// Attach a component to an entity. If a component with this type is already
// attached, this method will overwrite it.
func (db *EntityComponentDB) attach(entity uint64, component Attachable, index int) {
	if entity == 0 {
		log.Printf("Tried to attach 0 entity!")
		return
	}
	component.ResetRef()
	component.Ref().Entity = entity
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
			db.Components[index][componentsIndex] = component
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
	db.Components[index] = append(db.Components[index], component)
	component.SetIndexInDB(len(db.Components[index]) - 1)
}

// Create a new component with the given index and attach it.
func (db *EntityComponentDB) NewAttachedComponent(entity uint64, index int) Attachable {
	t := DbTypes().Types[index]
	newc := reflect.New(t).Interface()
	attached := newc.(Attachable)
	db.attach(entity, attached, index)
	attached.Construct(nil)
	return attached
}

func (db *EntityComponentDB) LoadAttachComponent(index int, data map[string]any, ignoreSerializedEntity bool) Attachable {
	var entity uint64
	var err error
	if ignoreSerializedEntity || data["Entity"] == nil {
		entity = db.NewEntity()
	} else if entity, err = strconv.ParseUint(data["Entity"].(string), 10, 64); entity == 0 || err != nil {
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
	component.ResetRef()
	component.Ref().Entity = 0
	component.SetDB(db)
	component.Construct(data)
	return component
}

func (db *EntityComponentDB) NewAttachedComponentTyped(entity uint64, cType string) Attachable {
	if index, ok := DbTypes().Indexes[cType]; ok {
		return db.NewAttachedComponent(entity, index)
	}

	log.Printf("NewComponent: unregistered type %v for entity %v\n", cType, entity)
	return nil
}

func (db *EntityComponentDB) Attach(componentIndex int, entity uint64, component Attachable) {
	db.attach(entity, component, componentIndex)
}

// This seems expensive. Need to profile
func (db *EntityComponentDB) AttachTyped(entity uint64, component Attachable) {
	t := reflect.ValueOf(component).Type()

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if index, ok := DbTypes().Indexes[t.String()]; ok {
		db.attach(entity, component, index)
	}
}

func (db *EntityComponentDB) Detach(index int, entity uint64) {
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
	components := db.Components[index]
	size := len(components)
	if size > i {
		components[i] = components[size-1]
		components[i].SetIndexInDB(i)
		db.Components[index] = components[:size-1]
	} else {
		log.Printf("EntityComponentDB.Detach: found entity %v component index %v, but component list is too short.", entity, index)
	}
	ec[index] = nil
}

func (db *EntityComponentDB) DetachByType(component Attachable) {
	if component == nil {
		return
	}
	entity := component.Ref().Entity

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

	component.ResetRef()
}

func (db *EntityComponentDB) DetachAll(entity uint64) {
	if entity == 0 {
		return
	}

	for index := range db.Components {
		db.Detach(index, entity)
	}

	db.EntityComponents[entity] = nil
}

func (db *EntityComponentDB) GetEntityRefByName(name string) *EntityRef {
	if allNamed := db.All(NamedComponentIndex); allNamed != nil {
		for _, c := range allNamed {
			named := c.(*Named)
			if named.Name == name {
				return named.EntityRef
			}
		}
	}
	return nil
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

func (db *EntityComponentDB) SerializeEntity(entity uint64) map[string]any {
	components := db.EntityComponents[entity]
	jsonEntity := make(map[string]any)
	jsonEntity["Entity"] = strconv.FormatUint(entity, 10)
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

	sortedEntities := make([]uint64, 0)
	for entity, c := range db.EntityComponents {
		if entity == 0 || c == nil {
			continue
		}
		sortedEntities = append(sortedEntities, uint64(entity))
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

func (db *EntityComponentDB) DeserializeEntityRefs(data []any) map[uint64]*EntityRef {
	result := make(map[uint64]*EntityRef)

	for _, v := range data {
		if entity, err := strconv.ParseUint(v.(string), 10, 64); err == nil {
			result[entity] = db.EntityRef(entity)
		}
	}
	return result
}

func (db *EntityComponentDB) DeserializeEntityRef(data any) *EntityRef {
	if data == nil {
		return nil
	}

	if entity, err := strconv.ParseUint(data.(string), 10, 64); err == nil {
		return db.EntityRef(entity)
	}
	return nil
}
