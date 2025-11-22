package ecs

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"tlyakhov/gofoom/concepts"
)

type Snapshot []any
type funcFileVistor func(entity Entity, data map[string]any) error

var attachedType = reflect.TypeFor[Attached]()

// TODO: We need to be able to test this somehow
func processNonSerializedFields(object any, serialized map[string]any, save bool) {
	if object == nil {
		return
	}

	objectValue := reflect.ValueOf(object).Elem()
	objectType := objectValue.Type()

	for i := range objectType.NumField() {
		field := objectType.Field(i)

		// Ignore editable, unexported, or specifically tagged fields.
		flags := FieldFlagsFromTag(field.Tag)
		shallowCache := (flags & FieldShallowCacheable) != 0
		isEditable := field.Tag.Get("editable") != ""
		isSliceOfStruct := field.Type.Kind() == reflect.Slice && isStructOrPtrToStruct(field.Type.Elem())
		isEditableNonStruct := (isEditable && field.Type.Kind() != reflect.Struct && !isSliceOfStruct)
		isAttached := field.Type == attachedType
		if !field.IsExported() || isEditableNonStruct || (flags&FieldNonCacheable) != 0 || isAttached {
			continue
		}
		// Shallow copy by default, except for embedded structs/value types
		v := objectValue.Field(i)
		name := "_cache_" + field.Name
		switch {
		case !shallowCache && field.Type.Kind() == reflect.Struct:
			child := v.Addr().Interface()
			childMap := make(map[string]any)
			if save {
				serialized[name] = childMap
			} else if loaded, ok := serialized[name]; ok {
				childMap = loaded.(map[string]any)
			}
			processNonSerializedFields(child, childMap, save)
		case !shallowCache && isSliceOfStruct:
			if save {
				childSlice := make([]map[string]any, v.Len())
				for i := range v.Len() {
					childSlice[i] = make(map[string]any)
					child := v.Index(i)
					processNonSerializedFields(ensurePointerToStruct(child), childSlice[i], true)
				}
				serialized[name] = childSlice
			} else if loaded, ok := serialized[name]; ok {
				loadedSlice := loaded.([]map[string]any)
				for v.Len() < len(loadedSlice) {
					v = reflect.Append(v, reflect.Zero(v.Type().Elem()))
					objectValue.Field(i).Set(v)
				}
				for i, childMap := range loadedSlice {
					child := v.Index(i)
					processNonSerializedFields(ensurePointerToStruct(child), childMap, false)
				}
			}
		case v.CanInterface():
			if save {
				serialized[name] = v.Interface()
			} else if loaded, ok := serialized[name]; ok {
				v.Set(reflect.ValueOf(loaded))
			}
		}
	}
}

func serializeEntity(entity Entity, visited map[uint64]Entity, includeNonSerialized bool) map[string]any {
	serialized := make(map[string]any)
	serialized["Entity"] = entity.Serialize()
	sid, local := entity.SourceID(), entity.Local()
	for _, component := range rows[sid][int(local)] {
		if component == nil || (!includeNonSerialized && component.Base().Flags&ComponentNoSave != 0) {
			continue
		}
		cid := component.ComponentID()
		hash := (uint64(component.Base().indexInArena) << 16) | (uint64(cid) & 0xFFFF)
		arena := Types().ArenaPlaceholders[cid]
		snapshotID := arena.Type().String()

		// If the caller is tracking serialized components and this component
		// is already saved, just reference the entity.
		if visited != nil {
			if savedEntity, ok := visited[hash]; ok {
				serialized[snapshotID] = savedEntity.Serialize()
				continue
			}
		}

		if !includeNonSerialized && component.Base().IsExternal() {
			// Just pick one
			// TODO: This has a code smell. Should there be a particular way
			// to pick an entity ID to reference when saving?
			serialized[snapshotID] = component.Base().ExternalEntities()[0].Serialize()
			continue
		}

		componentMap := component.Serialize()
		if includeNonSerialized {
			processNonSerializedFields(component, componentMap, true)
		}
		delete(componentMap, "Entities")
		serialized[snapshotID] = componentMap
		if visited != nil {
			visited[hash] = entity
		}

	}
	if len(serialized) == 1 {
		// No components were serialized, only the "Entity" field.
		return nil
	}
	return serialized
}

func visitSnapshotEntity(snapshotMap map[string]any, fn funcFileVistor) error {
	var snapshotEntity string
	var entity Entity
	var ok bool
	var err error

	if snapshotMap["Entity"] == nil {
		return errors.New("ecs.visitSnapshotEntity: snapshot map doesn't have entity key")
	}
	if snapshotEntity, ok = snapshotMap["Entity"].(string); !ok {
		return errors.New("ecs.visitSnapshotEntity: snapshot entity isn't string")
	}
	if entity, err = ParseEntity(snapshotEntity); err != nil {
		return fmt.Errorf("ecs.visitSnapshotEntity: snapshot entity can't be parsed: %w", err)
	}

	return fn(entity, snapshotMap)
}

func rangeSnapshot(snapshot Snapshot, fn funcFileVistor) error {
	for _, snapshotMap := range snapshot {
		snapshotEntity := snapshotMap.(map[string]any)
		if snapshotEntity == nil {
			log.Printf("ecs.rangeSnapshot: snapshot array element should be a map")
			continue
		}
		if err := visitSnapshotEntity(snapshotEntity, fn); err != nil {
			return err
		}
	}
	return nil
}

func SerializeEntity(entity Entity, includeNonSerialized bool) map[string]any {
	return serializeEntity(entity, nil, includeNonSerialized)
}

func SaveSnapshot(includeNonSerialized bool) Snapshot {
	defer concepts.ExecutionDuration(concepts.ExecutionTrack("ecs.SaveSnapshot"))
	Lock.Lock()
	defer Lock.Unlock()

	snapshot := Snapshot{}
	savedComponents := make(map[uint64]Entity)

	Entities.Range(func(entity uint32) {
		e := Entity(entity)
		if e == 0 {
			return
		}
		if !includeNonSerialized && e.IsExternal() {
			return
		}
		serialized := serializeEntity(e, savedComponents, includeNonSerialized)
		if len(serialized) == 0 {
			return
		}
		snapshot = append(snapshot, serialized)
	})
	return snapshot
}

func LoadSnapshot(snapshot Snapshot) error {
	defer concepts.ExecutionDuration(concepts.ExecutionTrack("ecs.LoadSnapshot"))

	Initialize()
	err := rangeSnapshot(snapshot, func(entity Entity, data map[string]any) error {
		Entities.Set(uint32(entity))

		for componentName, cid := range Types().IDs {
			componentData := data[componentName]
			if componentData == nil {
				continue
			}

			if linkedEntitySerialized, ok := componentData.(string); ok {
				linkedEntity, _ := ParseEntity(linkedEntitySerialized)
				if linkedEntity == 0 {
					continue
				}
				c := GetComponent(linkedEntity, cid)
				if c != nil {
					attach(entity, &c, cid)
				}
			} else {
				componentMap := componentData.(map[string]any)
				var attached Component
				attach(entity, &attached, cid)
				if attached.Base().Attachments == 1 {
					attached.Construct(componentMap)
					processNonSerializedFields(attached, componentMap, false)
				}
				if cid == SourceFileCID {
					file := attached.(*SourceFile)
					SourceFileNames[file.Source] = file
					SourceFileIDs[file.ID] = file
				}
			}
		}
		return nil
	})
	ActAllControllers(ControllerRecalculate)
	return err
}
