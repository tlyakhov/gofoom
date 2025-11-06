package ecs

import (
	"errors"
	"fmt"
	"log"
	"tlyakhov/gofoom/concepts"
)

type Snapshot []any

type funcFileVistor func(entity Entity, data map[string]any) error

func visitSnapshotEntity(snapshotMap map[string]any, fn funcFileVistor) error {
	var snapshotEntity string
	var entity Entity
	var ok bool
	var err error

	if snapshotMap["Entity"] == nil {
		return errors.New("ecs.visitSnapshotEntity: yaml object doesn't have entity key")
	}
	if snapshotEntity, ok = snapshotMap["Entity"].(string); !ok {
		return errors.New("ecs.visitSnapshotEntity: yaml entity isn't string")
	}
	if entity, err = ParseEntity(snapshotEntity); err != nil {
		return fmt.Errorf("ecs.visitSnapshotEntity: yaml entity can't be parsed: %w", err)
	}

	return fn(entity, snapshotMap)
}

func rangeSnapshot(snapshot Snapshot, fn funcFileVistor) error {
	for _, snapshotMap := range snapshot {
		snapshotEntity := snapshotMap.(map[string]any)
		if snapshotEntity == nil {
			log.Printf("ecs.rangeSnapshot: YAML array element should be an object")
			continue
		}
		if err := visitSnapshotEntity(snapshotEntity, fn); err != nil {
			return err
		}
	}
	return nil
}

func SerializeEntity(entity Entity, includeCaches bool) map[string]any {
	return serializeEntity(entity, nil, includeCaches)
}

func SaveSnapshot(includeCaches bool) Snapshot {
	defer concepts.ExecutionDuration(concepts.ExecutionTrack("ecs.Snapshot"))
	Lock.Lock()
	defer Lock.Unlock()

	snapshot := []any{}
	savedComponents := make(map[uint64]Entity)

	Entities.Range(func(entity uint32) {
		e := Entity(entity)
		if !includeCaches && e.IsExternal() {
			return
		}
		serialized := serializeEntity(e, savedComponents, includeCaches)
		if len(serialized) == 0 {
			return
		}
		snapshot = append(snapshot, serialized)
	})
	return Snapshot(snapshot)
}

func LoadSnapshot(snapshot Snapshot) error {
	// TODO: Redo this, don't follow the SourceFile hierarchy, just load direct
	defer concepts.ExecutionDuration(concepts.ExecutionTrack("ecs.LoadSnapshot"))

	Initialize()
	err := rangeSnapshot(snapshot, func(entity Entity, data map[string]any) error {
		Entities.Set(uint32(entity))

		for name, cid := range Types().IDs {
			yamlData := data[name]
			if yamlData == nil {
				continue
			}

			if yamlLink, ok := yamlData.(string); ok {
				linkedEntity, _ := ParseEntity(yamlLink)
				if linkedEntity == 0 {
					continue
				}
				c := GetComponent(linkedEntity, cid)
				if c != nil {
					attach(entity, &c, cid)
				}
			} else {
				yamlComponent := yamlData.(map[string]any)
				var attached Component
				attach(entity, &attached, cid)
				if attached.Base().Attachments == 1 {
					attached.Construct(yamlComponent)
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
	log.Printf("err: %v", err)
	return err
	/*
	   // Preserve the filename for SourceFile 0 if we have one
	   var primaryFile *SourceFile
	   arena := ArenaFor[SourceFile](SourceFileCID)

	   	for i := range arena.Cap() {
	   		file := arena.Value(i)
	   		if file == nil || file.ID != 0 {
	   			continue
	   		}
	   		primaryFile = &SourceFile{Source: file.Source}
	   		break
	   	}

	   	if primaryFile == nil {
	   		primaryFile = &SourceFile{}
	   	}

	   Initialize()
	   var a Component = primaryFile
	   Attach(SourceFileCID, NewEntity(), &a)
	   primaryFile = a.(*SourceFile)
	   primaryFile.ID = 0
	   primaryFile.Flags = ComponentInternal
	   primaryFile.SetContents(snapshot)
	   return primaryFile.Load()
	*/
}
