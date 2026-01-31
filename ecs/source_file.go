// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/pierrec/xxHash/xxHash32"
	"github.com/spf13/cast"
	"sigs.k8s.io/yaml"
)

type SourceFileHash uint32

type SourceFile struct {
	Attached `editable:"^"`

	Source     string `editable:"File" edit_type:"file"`
	ID         EntitySourceID
	Loaded     bool
	References int

	serializedContents Snapshot
	children           EntityTable
	loadedHash         SourceFileHash
	workingDir         string
}

var SourceFileCID ComponentID

func init() {
	SourceFileCID = RegisterComponent(&Arena[SourceFile, *SourceFile]{})
}

func (*SourceFile) ComponentID() ComponentID {
	return SourceFileCID
}

func GetSourceFile(e Entity) *SourceFile {
	if asserted, ok := GetComponent(e, SourceFileCID).(*SourceFile); ok {
		return asserted
	}
	return nil
}

func (file *SourceFile) SetContents(contents []any) {
	file.serializedContents = contents
}

func (file *SourceFile) Hash(forSerialization bool) SourceFileHash {
	if !forSerialization && file.loadedHash != 0 {
		return file.loadedHash
	}
	// TODO: Should we consider custom hashes?
	switch {
	case file.Source != "":
		x := xxHash32.New(0xCAFE)
		x.Write([]byte(file.Source))
		return SourceFileHash(x.Sum32())
	case file.Entity != 0:
		return SourceFileHash(file.Entity)
	default:
		return 0
	}
}

func (file *SourceFile) read(path string) error {
	// TODO: Streaming or lazy loads?
	bytes, err := os.ReadFile(path)

	if err != nil {
		return fmt.Errorf("SourceFile.read: reading file: %w", err)
	}

	contents := string(bytes)

	var yamlTree any
	if err := yaml.Unmarshal([]byte(contents), &yamlTree); err != nil {
		return fmt.Errorf("SourceFile.read: yaml parsing: %w", err)
	}

	var ok bool
	if file.serializedContents, ok = yamlTree.([]any); !ok || file.serializedContents == nil {
		return fmt.Errorf("SourceFile.read: YAML root must be an array")
	}

	return nil
}

var sourceFileTypeName = reflect.TypeFor[SourceFile]().String()

func (file *SourceFile) readAndMapNestedFiles(parent *SourceFile) error {
	/* 	The complexity in this code is due to handling the following case:
	   	Let's say we have a list of yaml files like this:

	   	1. basics.yaml - bunch of common entities like proximity scripts
	   	2. props.yaml - common game objects
	   	3. weapons.yaml - game weapons
	   	4. enemies.yaml - game enemies
	   	5. mission1.yaml - the first level

	   	Now, let's say 'props.yaml' refers to 'basics.yaml' (ID 1), and
	   	enemies.yaml refers to 'weapons.yaml' (ID 1) and 'basics.yaml' (ID 2).
	   	Then, 'mission1.yaml' refers to 'weapons.yaml' and 'enemies.yaml'.

	   	Now we have a big problem - we have references to the *same file* (basics)
	   	from 2 nested references, and they have different IDs (1 and 2 respectively).
	   	This means that when loading files, we need to first find all the nested
	   	file references, de-duplicate them, and map whatever ID we found to an
	   	open one (which may or may not be the same as what the original file had)

	   	Afterwards, we can load the rest of the entities & components,
	   	substituting the file IDs that we have mapped for all relations. */

	log.Printf("SourceFile.readAndMapNestedFiles - reading %v(%v): %v", file.Source, file.ID, file.workingDir)

	// Read in the file if we need to
	if file.serializedContents == nil {
		path := file.Source
		if parent != nil {
			path = filepath.Join(parent.workingDir, path)
		}
		if err := file.read(path); err != nil {
			return err
		}
	}

	if file.Entity != 1 {
		file.ID = NextFreeEntitySourceID()
	}
	file.References = 1
	SourceFileNames[file.Source] = file
	SourceFileIDs[file.ID] = file

	err := rangeSnapshot(file.serializedContents, func(entity Entity, data map[string]any) error {
		// Sanity check:
		if entity.IsExternal() {
			return fmt.Errorf("SourceFile.readAndMapNestedFiles: Warning, entity %v from file %v is external", entity, file.Source)
		}
		entity = entity.WithFileID(file.ID)
		// Only load SourceFile components
		yamlSourceFile := data[sourceFileTypeName] // ecs.SourceFile
		if yamlSourceFile == nil {
			return nil
		}
		data, ok := yamlSourceFile.(map[string]any)
		if !ok {
			return fmt.Errorf("SourceFile.readAndMapNestedFiles: Warning, couldn't parse ecs.SourceFile on entity %v", entity)
		}
		// Let's deserialize this reference and check if this file already
		// exists. If it does, we don't need to map it again.
		a := LoadComponentWithoutAttaching(SourceFileCID, data)
		nestedFile := a.(*SourceFile)
		if len(nestedFile.Source) == 0 {
			// Since the source is blank, maybe the user intends to fill
			// this in later. Let's attach it and not do any mapping.
			Entities.Set(uint32(entity))
			attach(entity, &a, SourceFileCID)
			return nil
		}
		if mappedFile, ok := SourceFileNames[nestedFile.Source]; ok {
			// This file is already mapped. We can skip loading it.
			file.children.Set(mappedFile.Entity)
			mappedFile.References++
			return nil
		}

		// If we're here, this file doesn't exist yet. Let's attach it and
		// map any nested files inside.
		Entities.Set(uint32(entity))
		attach(entity, &a, SourceFileCID)
		// the attach method will change where *a is.
		nestedFile = a.(*SourceFile)
		nestedFile.setWorkingDir(file.workingDir)
		nestedFile.readAndMapNestedFiles(file)
		file.children.Set(nestedFile.Entity)
		return nil
	})
	return err
}

func (file *SourceFile) setWorkingDir(parent string) {
	src := path.Dir(file.Source)
	if filepath.IsAbs(src) {
		var err error
		src, err = filepath.Rel(parent, src)
		if err != nil {
			log.Printf("SourceFile.setWorkingDir: error making relative path %v, %v", parent, src)
			src = path.Dir(file.Source)
		}
	}
	file.workingDir = filepath.Join(parent, src)
	if strings.HasSuffix(file.workingDir, "worlds") {
		// Only do this if our source matches expectations, to avoid unexpected path traversal.
		file.workingDir = filepath.Join(file.workingDir, "..")
	}
}

func (file *SourceFile) WorkingDir() string {
	return file.workingDir
}

func (file *SourceFile) Load() error {
	Lock.Lock()
	defer Lock.Unlock()

	// By convention, our working directory is one level below the world.
	wd, _ := os.Getwd()
	file.setWorkingDir(wd)
	file.Loaded = false
	if err := file.readAndMapNestedFiles(nil); err != nil {
		return err
	}
	err := file.loadEntities()
	file.serializedContents = nil
	// After everything's loaded, trigger the controllers
	ActAllControllers(ControllerPrecompute)
	return err
}

func (file *SourceFile) loadEntities() error {
	file.Loaded = true
	// Go depth first so that references resolve right away.
	for _, e := range file.children {
		nestedFile := GetSourceFile(e)
		if nestedFile == nil {
			continue
		}
		if nestedFile.Loaded {
			continue
		}
		nestedFile.loadEntities()
	}
	err := rangeSnapshot(file.serializedContents, func(entity Entity, data map[string]any) error {
		// This file has an ID assigned to it, so every entity and component in
		// the file has to include this file ID. We also need to map any
		// relations.
		entity = entity.WithFileID(file.ID)
		Entities.Set(uint32(entity))

		for name, cid := range Types().IDs {
			if cid == SourceFileCID {
				// We've already attached SourceFiles in readAndMapNestedFiles
				continue
			}
			yamlData := data[name]
			if yamlData == nil {
				continue
			}

			if yamlLink, ok := yamlData.(string); ok {
				linkedEntity, _ := ParseEntity(yamlLink)
				if linkedEntity == 0 {
					continue
				}
				if !linkedEntity.IsExternal() {
					linkedEntity = linkedEntity.WithFileID(file.ID)
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
				ModifyComponentRelationEntities(attached, func(r *Relation, e Entity) Entity {
					if e == 0 {
						return 0
					}
					if !e.IsExternal() {
						return e.WithFileID(file.ID)
					}
					return e
				})
			}
		}
		return nil
	})
	return err
}

func (file *SourceFile) Unload() {
	if file.ID == 0 || !file.Loaded {
		return
	}

	// First unload any children
	for _, e := range file.children {
		nestedFile := GetSourceFile(e)
		if nestedFile == nil {
			continue
		}
		nestedFile.References--
		if nestedFile.References == 0 {
			nestedFile.Unload()
		}
	}

	// We need to detach/delete all the entities from this file and unmap this ID.
	// References will end up broken unless cleaned up beforehand.
	toDelete := make([]Entity, 0)
	Entities.Range(func(entity uint32) {
		e := Entity(entity)
		if e.SourceID() != file.ID {
			return
		}
		toDelete = append(toDelete, e)
	})
	for _, e := range toDelete {
		Delete(e)
	}

	delete(SourceFileNames, file.Source)
	delete(SourceFileIDs, file.ID)
	file.Loaded = false
	file.serializedContents = nil
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
	file.Loaded = false
	file.References = 0
	file.serializedContents = nil
	file.children = EntityTable{}

	if data == nil {
		return
	}

	if v, ok := data["Source"]; ok {
		file.Source = v.(string)
	}
	if v, ok := data["ID"]; ok {
		file.ID = EntitySourceID(cast.ToUint8(v))
	}
	if v, ok := data["Hash"]; ok {
		file.loadedHash = SourceFileHash(cast.ToUint32(v))
	}
}

func (file *SourceFile) Serialize() map[string]any {
	result := file.Attached.Serialize()

	result["Source"] = file.Source
	result["ID"] = file.ID
	result["Hash"] = "0x" + strconv.FormatUint(uint64(file.Hash(true)), 16) // To handle file moves and human readers

	return result
}

func SourceFileFromHash(hash SourceFileHash) *SourceFile {
	arena := ArenaFor[SourceFile](SourceFileCID)
	for i := range arena.Cap() {
		file := arena.Value(i)
		if file == nil {
			continue
		}
		if file.Hash(true) == hash {
			return file
		} else if file.loadedHash == hash {
			log.Printf("Warning: used loaded file hash %x when parsing entity.", hash)
			return file
		}
	}
	return nil
}

func SourceFileForEntity(e Entity) *SourceFile {
	if file, ok := SourceFileIDs[e.SourceID()]; ok {
		return file
	}
	return nil
}

func WorkingDirForEntity(e Entity) string {
	file := SourceFileForEntity(e)
	if file == nil {
		// Maybe unattached? in that case, try the topmost sourcefile
		if file, ok := SourceFileIDs[0]; ok {
			return file.workingDir
		}
		// No luck? return empty
		return ""
	}
	return file.workingDir
}
