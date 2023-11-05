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
	"sync/atomic"
)

// The architecture is like this:
// * An entity is a globally unique uint64, e.g. primary key
// * A component is a named (string) table with columns of fields, rows of
// entities
// * A system is code that queries and operates on components and entities
type EntityComponentDB struct {
	*Simulation
	Components       []map[uint64]Attachable
	EntityComponents map[uint64][]Attachable
	currentEntity    uint64
	lock             sync.RWMutex
}

func NewEntityComponentDB() *EntityComponentDB {
	db := &EntityComponentDB{}
	db.Clear()

	return db
}

func (db *EntityComponentDB) Clear() {
	db.currentEntity = 0
	db.EntityComponents = make(map[uint64][]Attachable)
	db.Components = make([]map[uint64]Attachable, len(DbTypes().Indexes))
	db.Simulation = NewSimulation()
	for i := 0; i < len(DbTypes().Indexes); i++ {
		db.Components[i] = make(map[uint64]Attachable)
	}
}

func (db *EntityComponentDB) NewEntity() uint64 {
	return atomic.AddUint64(&db.currentEntity, 1)
}

func (db *EntityComponentDB) NewEntityRef() *EntityRef {
	return &EntityRef{Entity: db.NewEntity(), DB: db}
}

func (db *EntityComponentDB) EntityRef(entity uint64) *EntityRef {
	return &EntityRef{DB: db, Entity: entity}
}

func (db *EntityComponentDB) All(index int) map[uint64]Attachable {
	return db.Components[index]
}

func (db *EntityComponentDB) AllForType(cType string) map[uint64]Attachable {
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

func (db *EntityComponentDB) attach(entity uint64, component Attachable, index int) {
	if entity == 0 {
		log.Printf("Tried to attach 0 entity!")
	}
	component.Ref().Reset()
	component.Ref().Entity = entity
	component.SetDB(db)
	db.Components[index][entity] = component
	if ec, ok := db.EntityComponents[entity]; ok {
		ec[index] = component
	} else {
		ec := make([]Attachable, len(DbTypes().Types))
		ec[index] = component
		db.EntityComponents[entity] = ec
	}
}

func (db *EntityComponentDB) NewComponent(entity uint64, index int) Attachable {
	t := DbTypes().Types[index]
	newc := reflect.New(t).Interface()
	attached := newc.(Attachable)
	db.attach(entity, attached, index)
	attached.Construct(nil)
	return attached
}

func (db *EntityComponentDB) LoadComponent(index int, data map[string]any) Attachable {
	var entity uint64
	var err error
	if data["Entity"] == nil {
		entity = db.NewEntity()
	} else if entity, err = strconv.ParseUint(data["Entity"].(string), 10, 64); entity == 0 || err != nil {
		entity = db.NewEntity()
	}

	if db.currentEntity < entity {
		db.currentEntity = entity
	}

	t := DbTypes().Types[index]
	newc := reflect.New(t).Interface()
	attached := newc.(Attachable)
	db.attach(entity, attached, index)
	attached.Construct(data)
	return attached
}

func (db *EntityComponentDB) NewComponentTyped(entity uint64, cType string) Attachable {
	if index, ok := DbTypes().Indexes[cType]; ok {
		return db.NewComponent(entity, index)
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
		return
	}

	delete(db.Components[index], entity)
	if ec, ok := db.EntityComponents[entity]; ok {
		ec[index] = nil
	}
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

	component.Ref().Reset()
}

func (db *EntityComponentDB) DetachAll(entity uint64) {
	if entity == 0 {
		return
	}

	for index := range db.Components {
		delete(db.Components[index], entity)
	}

	delete(db.EntityComponents, entity)
}

func (db *EntityComponentDB) GetEntityRefByName(name string) *EntityRef {
	if allNamed := db.All(NamedComponentIndex); allNamed != nil {
		for entity, c := range allNamed {
			named := c.(*Named)
			if named.Name == name {
				return db.EntityRef(entity)
			}
		}
	}
	return &EntityRef{}
}

func (db *EntityComponentDB) DeserializeEntity(jsonEntity map[string]any) {
	for name, index := range DbTypes().Indexes {
		jsonData := jsonEntity[name]
		if jsonData == nil {
			continue
		}
		jsonComponent := jsonData.(map[string]any)
		db.LoadComponent(index, jsonComponent)
	}
}
func (db *EntityComponentDB) Load(filename string) error {
	db.lock.Lock()
	defer db.lock.Unlock()

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
		db.DeserializeEntity(jsonEntity)
	}

	// After everything's loaded, trigger the controllers
	set := db.NewControllerSet()
	set.ActGlobal(ControllerRecalculate)
	set.ActGlobal(ControllerLoaded)
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
	db.lock.Lock()
	defer db.lock.Unlock()
	jsonDB := make([]any, 0)

	sortedEntities := make([]uint64, len(db.EntityComponents))
	i := 0
	for entity := range db.EntityComponents {
		sortedEntities[i] = entity
		i++
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
