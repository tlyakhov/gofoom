// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cast"
	"gopkg.in/yaml.v3"
)

type SourceFile struct {
	Attached `editable:"^"`

	Source     string `editable:"File" edit_type:"file"`
	ID         EntitySourceID
	OriginalID EntitySourceID
	Loaded     bool
}

var SourceFileCID ComponentID

func init() {
	SourceFileCID = RegisterComponent(&Column[SourceFile, *SourceFile]{Getter: GetSourceFile})
}

func GetSourceFile(db *ECS, e Entity) *SourceFile {
	if asserted, ok := db.Component(e, SourceFileCID).(*SourceFile); ok {
		return asserted
	}
	return nil
}

type processFileEntitiesFunc func(builder *strings.Builder, e Entity, name string, file string) bool

func iterateFileEntities(contents string, fn processFileEntitiesFunc) string {
	builder := strings.Builder{}
	builder.Grow(len(contents))

	prefixRunes := []rune(EntityDelimiter)
	limit := len(contents) - entityDelimiterLength
	skip := 0
	// Look for entity strings, parse them, call `fn`.
	for i, r := range contents {
		if skip > 0 {
			skip--
			continue
		}
		if i >= limit || r != prefixRunes[0] {
			builder.WriteRune(r)
			continue
		}

		parts := EntityRegexp.FindStringSubmatch(contents[i:])
		if parts == nil {
			log.Printf("Can't parse entity " + contents[i:i+3])
			builder.WriteRune(r)
			continue
		}
		v, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			log.Printf("SourceFile.replaceSourceID: error parsing entity %v", contents[i:i+3])
			builder.WriteRune(r)
			continue
		}
		entity := Entity(v)
		name := ""
		file := ""
		if len(parts) > 2 {
			name, _ = url.QueryUnescape(parts[2])
		}
		if len(parts) > 3 {
			file, _ = url.QueryUnescape(parts[3])
		}

		if !fn(&builder, entity, name, file) {
			builder.WriteString(parts[0])
		}
		skip = len([]rune(parts[0])) - 1
	}

	fmt.Println(builder.String())
	return builder.String()
}

func (file *SourceFile) visitFileEntity(yamlMap map[string]any, fn func(entity Entity, data map[string]any)) {
	var yamlEntity string
	var entity Entity
	var ok bool
	var err error
	if yamlMap["Entity"] == nil {
		log.Printf("SourceFile.visitFileEntity: yaml object doesn't have entity key")
		return
	}
	if yamlEntity, ok = yamlMap["Entity"].(string); !ok {
		log.Printf("SourceFile.visitFileEntity: yaml entity isn't string")
		return
	}
	if entity, err = ParseEntity(yamlEntity); err != nil {
		log.Printf("SourceFile.visitFileEntity: yaml entity can't be parsed: %v", err)
		return
	}

	fn(entity, yamlMap)
}

func (file *SourceFile) rangeFile(contents string, fn func(entity Entity, data map[string]any)) error {
	var parsed any
	err := yaml.Unmarshal([]byte(contents), &parsed)
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
		file.visitFileEntity(yamlEntity, fn)
	}
	return nil
}

func (file *SourceFile) mapFile() (string, error) {
	// TODO: there's bugs here if the user tries to map a file with the same
	// source filename.

	// Check if the source file ID we're loading is free or if it's
	// already mapped. If it's mapped, we have to pick a new one.
	file.OriginalID = file.ID
	file.Loaded = false
	if _, ok := file.ECS.SourceFileIDs[file.ID]; ok {
		file.ID = file.ECS.NextFreeEntitySourceID()
	}
	file.ECS.SourceFileNames[file.Source] = file
	file.ECS.SourceFileIDs[file.ID] = file

	contents, err := file.read()
	if err != nil {
		log.Printf("SourceFile.mapFile: error loading file %v: %v", file.Source, err)
		return contents, err
	}

	contentsSubbed := iterateFileEntities(contents,
		func(builder *strings.Builder, e Entity, name, filename string) bool {
			// Skip any external entities, we only care about SourceFiles,
			// so this should be safe, despite the fact that we may have
			// ID overlaps in nested files.
			if e.IsExternal() {
				return false
			}
			e += Entity(file.ID) << EntityBits
			builder.WriteString(e.SerializeRaw(name, filename))
			return true
		})
	file.rangeFile(contentsSubbed, func(entity Entity, data map[string]any) {
		yamlSourceFile := data["ecs.SourceFile"]
		if yamlSourceFile == nil {
			return
		}
		if data, ok := yamlSourceFile.(map[string]any); ok {
			// Let's deserialize this reference and check if this file already
			// exists in the ECS. If it does, we don't need to map it again.
			a := file.ECS.LoadComponentWithoutAttaching(SourceFileCID, data)
			nfile := a.(*SourceFile)
			if len(nfile.Source) == 0 {
				// Since the source is blank, maybe the user intends to fill
				// this in later. Let's attach it and not do any mapping.
				file.ECS.attach(entity, &a, SourceFileCID)
				return
			}
			if _, ok := file.ECS.SourceFileNames[nfile.Source]; ok {
				// This file is already mapped. We can skip it.

				// TODO: We need to refcount the dependency, to ensure we don't
				// delete it when unloading
				return
			}

			// If we're here, this file doesn't exist in the ECS yet. Let's map
			// it and attach it.
			nfile.ECS.attach(entity, &a, SourceFileCID)
			// the attach method will change where *a is.
			nfile = a.(*SourceFile)
			nfile.mapFile()
		}
	})
	return contents, nil
}

func (file *SourceFile) loadAllNestedFiles(contents string) error {
	contents = iterateFileEntities(contents,
		func(builder *strings.Builder, e Entity, name, filename string) bool {
			// This entity is a member of the file we're processing, which may
			// have an ID that we need to substitute in
			if !e.IsExternal() {
				e += Entity(file.ID) << EntityBits
				builder.WriteString(e.SerializeRaw(name, file.Source))
				return true
			}

			// If we're here, we've got an external entity inside of a nested
			// file. We need to remove its source ID and add whichever one we've
			// already mapped. We could follow the EntitySourceID for this
			// entity to its SourceFile, but for simplicity we store the
			// filenames along with the entities.
			if len(filename) == 0 {
				// This should probably break loading of a file, since it could
				// corrupt our entity tables.
				log.Printf("SourceFile.loadAllNestedFiles: %v,%v,%v is external and has no filename", e, name, filename)
				return false
			}
			nestedFile := file.ECS.SourceFileNames[filename]
			e = Entity(nestedFile.ID)<<EntityBits + e&MaxEntities
			builder.WriteString(e.SerializeRaw(name, filename))
			return true
		})

	file.Loaded = true
	err := file.rangeFile(contents, func(entity Entity, data map[string]any) {
		file.ECS.Entities.Set(uint32(entity))

		for name, cid := range Types().IDs {
			yamlData := data[name]
			if yamlData == nil {
				continue
			}
			if cid == SourceFileCID {
				// we've already mapped them in mapAllNestedFiles, now we need
				// to actually load the file data.
				nestedFilename := cast.ToString(yamlData.(map[string]any)["Source"])
				if len(nestedFilename) == 0 {
					continue
				}
				nfile := file.ECS.SourceFileNames[nestedFilename]
				// Let's not do it twice!
				if nfile.Loaded {
					continue
				}
				nestedContents, err := nfile.read()
				if err != nil {
					log.Printf("SourceFile.loadAllNestedFiles: error reading file %v", nestedFilename)
					continue
				}
				nfile.loadAllNestedFiles(nestedContents)
				continue
			}

			if yamlLink, ok := yamlData.(string); ok {
				linkedEntity, _ := ParseEntity(yamlLink)
				if linkedEntity != 0 {
					c := file.ECS.Component(linkedEntity, cid)
					if c != nil {
						file.ECS.attach(entity, &c, cid)
					}
				}
			} else {
				yamlComponent := yamlData.(map[string]any)
				var attached Attachable
				file.ECS.attach(entity, &attached, cid)
				if attached.Base().Attachments == 1 {
					attached.Construct(yamlComponent)
				}
			}
		}
	})
	return err
}

func (file *SourceFile) read() (string, error) {
	// TODO: Streaming or lazy loads?
	bytes, err := os.ReadFile(file.Source)
	return string(bytes), err
}

func (file *SourceFile) Load() error {
	file.ECS.Lock.Lock()
	defer file.ECS.Lock.Unlock()

	if contents, err := file.mapFile(); err == nil {
		fmt.Println(contents)
		err = file.loadAllNestedFiles(contents)
		if err != nil {
			return err
		}
	}

	// After everything's loaded, trigger the controllers
	file.ECS.ActAllControllers(ControllerRecalculate)
	return nil
}

func (file *SourceFile) Unload() {
	if file.ECS == nil || file.ID == 0 || !file.Loaded {
		return
	}

	file.ECS.Lock.Lock()
	defer file.ECS.Lock.Unlock()

	// We need to detach/delete all the entities from this file and unmap this ID.
	// References will end up broken unless cleaned up beforehand.
	toDelete := make([]Entity, 0)
	file.ECS.Entities.Range(func(entity uint32) {
		e := Entity(entity)
		if !e.IsExternal() || e.SourceID() != file.ID {
			return
		}
		toDelete = append(toDelete, e)
	})
	for _, e := range toDelete {
		file.ECS.Delete(e)
	}

	delete(file.ECS.SourceFileNames, file.Source)
	delete(file.ECS.SourceFileIDs, file.ID)
}

func (file *SourceFile) String() string {
	return fmt.Sprintf("Source file (%v): %v", file.ID, file.Source)
}

func (file *SourceFile) OnDetach(e Entity) {
	defer file.Attached.OnDetach(e)

	file.Unload()
}

func (file *SourceFile) Construct(data map[string]any) {
	file.Attached.Construct(data)
	file.ID = 0
	file.OriginalID = 0xFF
	file.Loaded = false

	if data == nil {
		return
	}

	if v, ok := data["Source"]; ok {
		file.Source = v.(string)
	}

	if v, ok := data["ID"]; ok {
		file.ID = EntitySourceID(cast.ToUint8(v))
	}
}

func (file *SourceFile) Serialize() map[string]any {
	result := file.Attached.Serialize()

	result["Source"] = file.Source
	result["ID"] = file.ID
	return result
}
