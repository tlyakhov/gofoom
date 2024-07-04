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

func jsonNameToEntity(db *concepts.EntityComponentDB, jsonData map[string]any, key string, obj any) {
	log.Printf("[%v] Linking field %v...", reflect.TypeOf(obj).String(), key)
	if name, ok := jsonData[key]; ok {
		entity := db.GetEntityByName(name.(string))
		if entity != 0 {
			log.Printf("Found entity %v", entity.String(db))
			reflect.ValueOf(obj).Elem().FieldByName(key).Set(reflect.ValueOf(entity))
		}
	}
}

func convertBodies(db *concepts.EntityComponentDB, sector *core.Sector, jsonBodies []any) {
	for _, jsonData := range jsonBodies {
		jsonBody := jsonData.(map[string]any)
		bodyType := jsonBody["Type"].(string)
		log.Printf("Converting body %v", jsonBody["Name"])
		eBody := db.NewEntity()
		db.NewAttachedComponent(eBody, concepts.NamedComponentIndex)
		sector.Bodies[eBody] = db.NewAttachedComponent(eBody, core.BodyComponentIndex).(*core.Body)
		switch bodyType {
		case "mobs.Light":
			db.NewAttachedComponent(eBody, core.LightComponentIndex)
		}
		for index, c := range db.AllComponents(eBody) {
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
	eSpawn := db.NewEntity()
	spawn := db.NewAttachedComponent(eSpawn, core.SpawnComponentIndex).(*core.Spawn)
	spawn.Spawn[0] = json["SpawnX"].(float64)
	spawn.Spawn[1] = json["SpawnY"].(float64)

	// Materials
	jsonMaterials := json["Materials"].([]any)
	for _, jsonData := range jsonMaterials {
		jsonMaterial := jsonData.(map[string]any)
		log.Printf("Converting material %v", jsonMaterial["Name"])
		eMaterial := db.NewEntity()
		matType := jsonMaterial["Type"].(string)
		db.NewAttachedComponent(eMaterial, concepts.NamedComponentIndex)
		db.NewAttachedComponent(eMaterial, materials.ImageComponentIndex)
		switch matType {
		case "materials.LitSampled":
			db.NewAttachedComponent(eMaterial, materials.LitComponentIndex)
		case "materials.PainfulLitSampled":
			db.NewAttachedComponent(eMaterial, materials.LitComponentIndex)
			//db.NewComponent(imat.Entity, behaviors.ToxicComponentIndex)
		case "materials.Sky":
			//			db.NewComponent(imat.Entity, materials.SkyComponentIndex)
		}
		for index, c := range db.AllComponents(eMaterial) {
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
	sectorIndexes := make(map[int]concepts.Entity)
	jsonSectors := json["Sectors"].([]any)
	for index, jsonData := range jsonSectors {
		jsonSector := jsonData.(map[string]any)
		log.Printf("Converting sector %v", jsonSector["Name"])
		eSector := db.NewEntity()
		sectorIndexes[index] = eSector
		sectorType := jsonSector["Type"].(string)
		db.NewAttachedComponent(eSector, concepts.NamedComponentIndex)
		db.NewAttachedComponent(eSector, core.SectorComponentIndex)
		switch sectorType {
		case "sectors.VerticalDoor":
			db.NewAttachedComponent(eSector, behaviors.VerticalDoorComponentIndex)
		case "sectors.ToxicSector":
			//db.NewComponent(isector.Entity, behaviors.ToxicComponentIndex)
		}
		for _, c := range db.AllComponents(eSector) {
			if c == nil {
				continue
			}
			c.Construct(jsonSector)
			switch target := c.(type) {
			case *core.Sector:
				jsonNameToEntity(db, jsonSector, "CeilMaterial", target)
				jsonNameToEntity(db, jsonSector, "FloorMaterial", target)
				if jsonSector["Mobs"] != nil {
					convertBodies(db, target, jsonSector["Mobs"].([]any))
				}
			}
		}
	}
	for index, jsonData := range jsonSectors {
		jsonSector := jsonData.(map[string]any)
		log.Printf("Wiring up adjacent sectors %v", jsonSector["Name"])
		eSector := sectorIndexes[index]

		for _, c := range db.AllComponents(eSector) {
			if c == nil {
				continue
			}
			switch target := c.(type) {
			case *core.Sector:
				jsonNameToEntity(db, jsonSector, "CeilTarget", target)
				jsonNameToEntity(db, jsonSector, "FloorTarget", target)
				jsonSegments := jsonSector["Segments"].([]any)
				for i, seg := range target.Segments {
					jsonNameToEntity(db, jsonSegments[i].(map[string]any), "AdjacentSector", seg)
					jsonNameToEntity(db, jsonSegments[i].(map[string]any), "LoMaterial", seg)
					jsonNameToEntity(db, jsonSegments[i].(map[string]any), "MidMaterial", seg)
					jsonNameToEntity(db, jsonSegments[i].(map[string]any), "HiMaterial", seg)
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
