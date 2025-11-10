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

func SerializeEntity(entity Entity, includeCaches bool) map[string]any {
	return serializeEntity(entity, nil, includeCaches)
}

func SaveSnapshot(includeCaches bool) Snapshot {
	defer concepts.ExecutionDuration(concepts.ExecutionTrack("ecs.SaveSnapshot"))
	Lock.Lock()
	defer Lock.Unlock()

	snapshot := Snapshot{}
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
	return snapshot
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
	return err
}
