// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"
)

type flagsSlice []string

func (flags *flagsSlice) String() string {
	return strings.Join(*flags, ", ")
}

func (flags *flagsSlice) Set(value string) error {
	*flags = append(*flags, value)
	return nil
}

var inputFiles flagsSlice

func jsonRefToEntityRef(db *concepts.EntityComponentDB, jsonData map[string]any, key string, obj any) {
	log.Printf("[%v] Linking field %v...", reflect.TypeOf(obj).String(), key)
	if name, ok := jsonData[key]; ok {
		er := db.GetEntityRefByName(name.(string))
		if !er.Nil() {
			log.Printf("Found entity %v", er.String())
			reflect.ValueOf(obj).Elem().FieldByName(key).Set(reflect.ValueOf(er))
		}
	}
}

func convertBodies(db *concepts.EntityComponentDB, sector *core.Sector, jsonBodies []any) {
	for _, jsonData := range jsonBodies {
		jsonBody := jsonData.(map[string]any)
		bodyType := jsonBody["Type"].(string)
		log.Printf("Converting body %v", jsonBody["Name"])
		ibody := db.RefForNewEntity()
		sector.Bodies[ibody.Entity] = ibody
		db.NewAttachedComponent(ibody.Entity, concepts.NamedComponentIndex)
		db.NewAttachedComponent(ibody.Entity, core.BodyComponentIndex)
		switch bodyType {
		case "mobs.Light":
			db.NewAttachedComponent(ibody.Entity, core.LightComponentIndex)
		}
		for index, c := range ibody.All() {
			if c == nil {
				continue
			}
			if index != core.LightComponentIndex {
				c.Construct(jsonBody)
			} else {
				if b, ok := jsonBody["Behaviors"].([]any); ok && len(b) > 0 {
					c.Construct(b[0].(map[string]any))
				}
			}
		}
	}
}

func convert(filename, output string) {
	db := concepts.NewEntityComponentDB()

	fileContents, err := os.ReadFile(filename)

	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	var parsed any
	err = json.Unmarshal(fileContents, &parsed)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	json := parsed.(map[string]any)

	// Spawn
	ispawn := db.RefForNewEntity()
	spawn := db.NewAttachedComponent(ispawn.Entity, core.SpawnComponentIndex).(*core.Spawn)
	spawn.Spawn[0] = json["SpawnX"].(float64)
	spawn.Spawn[1] = json["SpawnY"].(float64)

	// Materials
	jsonMaterials := json["Materials"].([]any)
	for _, jsonData := range jsonMaterials {
		jsonMaterial := jsonData.(map[string]any)
		log.Printf("Converting material %v", jsonMaterial["Name"])
		imat := db.RefForNewEntity()
		matType := jsonMaterial["Type"].(string)
		db.NewAttachedComponent(imat.Entity, concepts.NamedComponentIndex)
		db.NewAttachedComponent(imat.Entity, materials.ImageComponentIndex)
		switch matType {
		case "materials.LitSampled":
			db.NewAttachedComponent(imat.Entity, materials.LitComponentIndex)
		case "materials.PainfulLitSampled":
			db.NewAttachedComponent(imat.Entity, materials.LitComponentIndex)
			//db.NewComponent(imat.Entity, behaviors.ToxicComponentIndex)
		case "materials.Sky":
			//			db.NewComponent(imat.Entity, materials.SkyComponentIndex)
		}
		for index, c := range imat.All() {
			if c == nil {
				continue
			}
			if index != materials.ImageComponentIndex {
				c.Construct(jsonMaterial)
			} else {
				c.Construct(jsonMaterial["Texture"].(map[string]any))
			}
		}
	}

	// Sectors & bodies
	sectorIndexes := make(map[int]uint64)
	jsonSectors := json["Sectors"].([]any)
	for index, jsonData := range jsonSectors {
		jsonSector := jsonData.(map[string]any)
		log.Printf("Converting sector %v", jsonSector["Name"])
		isector := db.RefForNewEntity()
		sectorIndexes[index] = isector.Entity
		sectorType := jsonSector["Type"].(string)
		db.NewAttachedComponent(isector.Entity, concepts.NamedComponentIndex)
		db.NewAttachedComponent(isector.Entity, core.SectorComponentIndex)
		switch sectorType {
		case "sectors.VerticalDoor":
			db.NewAttachedComponent(isector.Entity, behaviors.VerticalDoorComponentIndex)
		case "sectors.ToxicSector":
			//db.NewComponent(isector.Entity, behaviors.ToxicComponentIndex)
		}
		for _, c := range isector.All() {
			if c == nil {
				continue
			}
			c.Construct(jsonSector)
			switch target := c.(type) {
			case *core.Sector:
				jsonRefToEntityRef(db, jsonSector, "CeilMaterial", target)
				jsonRefToEntityRef(db, jsonSector, "FloorMaterial", target)
				if jsonSector["Mobs"] != nil {
					convertBodies(db, target, jsonSector["Mobs"].([]any))
				}
			}
		}
	}
	for index, jsonData := range jsonSectors {
		jsonSector := jsonData.(map[string]any)
		log.Printf("Wiring up adjacent sectors %v", jsonSector["Name"])
		isector := &concepts.EntityRef{DB: db, Entity: sectorIndexes[index]}

		for _, c := range isector.All() {
			if c == nil {
				continue
			}
			switch target := c.(type) {
			case *core.Sector:
				jsonRefToEntityRef(db, jsonSector, "CeilTarget", target)
				jsonRefToEntityRef(db, jsonSector, "FloorTarget", target)
				jsonSegments := jsonSector["Segments"].([]any)
				for i, seg := range target.Segments {
					jsonRefToEntityRef(db, jsonSegments[i].(map[string]any), "AdjacentSector", seg)
					jsonRefToEntityRef(db, jsonSegments[i].(map[string]any), "LoMaterial", seg)
					jsonRefToEntityRef(db, jsonSegments[i].(map[string]any), "MidMaterial", seg)
					jsonRefToEntityRef(db, jsonSegments[i].(map[string]any), "HiMaterial", seg)
				}
			}
		}
	}

	db.Save(output)
}

func main() {
	flag.Var(&inputFiles, "input", "Worlds to convert")
	flag.Parse()

	for _, flag := range inputFiles {
		if matches, err := filepath.Glob(flag); err == nil {
			for _, file := range matches {
				log.Printf("Converting %v", file)
				convert(file, strings.Replace(file, "oldworlds", "worlds", -1))
			}
		}
	}
}
