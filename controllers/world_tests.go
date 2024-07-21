// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"fmt"
	"log"
	"math/rand"
	"tlyakhov/gofoom/archetypes"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"
)

func CreateTestSector(db *concepts.EntityComponentDB, name string, x, y, size float64) *core.Sector {
	eSector := archetypes.CreateSector(db)
	sector := core.SectorFromDb(db, eSector)
	sector.Construct(nil)
	named := db.NewAttachedComponent(eSector, concepts.NamedComponentIndex).(*concepts.Named)
	named.Name = name

	mat := DefaultMaterial(db)
	sector.FloorSurface.Material = mat
	sector.CeilSurface.Material = mat
	seg := sector.AddSegment(x, y)
	seg.Surface.Material = mat
	seg.HiSurface.Material = mat
	seg.LoSurface.Material = mat
	seg = sector.AddSegment(x+size, y)
	seg.Surface.Material = mat
	seg.HiSurface.Material = mat
	seg.LoSurface.Material = mat
	seg = sector.AddSegment(x+size, y+size)
	seg.Surface.Material = mat
	seg.HiSurface.Material = mat
	seg.LoSurface.Material = mat
	seg = sector.AddSegment(x, y+size)
	seg.Surface.Material = mat
	seg.HiSurface.Material = mat
	seg.LoSurface.Material = mat

	sector.Recalculate()
	return sector
}

func CreateTestGrass(db *concepts.EntityComponentDB) concepts.Entity {
	eGrass := archetypes.CreateBasicMaterial(db, true)
	nmat := db.NewAttachedComponent(eGrass, concepts.NamedComponentIndex).(*concepts.Named)
	nmat.Name = "Default Material"
	//tex.Diffuse = color.NRGBA{R: 128, G: 100, B: 50, A: 255}
	tex := materials.ImageFromDb(db, eGrass)
	tex.Source = "data/grass.jpg"
	tex.Filter = true
	tex.GenerateMipMaps = true
	tex.Load()
	return eGrass
}
func CreateTestSky(db *concepts.EntityComponentDB) concepts.Entity {
	img := db.NewAttachedComponent(db.NewEntity(), materials.ImageComponentIndex).(*materials.Image)
	img.Source = "data/Sky.png"
	img.Filter = false
	img.GenerateMipMaps = false
	img.Load()

	entity := archetypes.CreateBasic(db, materials.ShaderComponentIndex)
	sky := materials.ShaderFromDb(db, entity)
	sky.Stages = append(sky.Stages, new(materials.ShaderStage))
	sky.Stages[0].Construct(nil)
	sky.Stages[0].Texture = img.Entity
	sky.Stages[0].Flags = materials.ShaderSky | materials.ShaderTiled
	named := db.NewAttachedComponent(entity, concepts.NamedComponentIndex).(*concepts.Named)
	named.Name = "Sky"

	return entity
}

func CreateTestDirt(db *concepts.EntityComponentDB) concepts.Entity {
	eDirt := archetypes.CreateBasicMaterial(db, true)
	nmat := db.NewAttachedComponent(eDirt, concepts.NamedComponentIndex).(*concepts.Named)
	nmat.Name = "Dirt"
	tex := materials.ImageFromDb(db, eDirt)
	tex.Source = "data/FDef.png"
	tex.Filter = false
	tex.GenerateMipMaps = true
	tex.Load()
	return eDirt
}

func CreateTestWorld(db *concepts.EntityComponentDB) {
	testw := 30
	testh := 30
	eSpawn := archetypes.CreateBasic(db, core.SpawnComponentIndex)
	spawn := core.SpawnFromDb(db, eSpawn)
	spawn.Spawn[0] = 250
	spawn.Spawn[1] = 250
	spawn.Spawn[2] = 100

	eGrass := archetypes.CreateBasicMaterial(db, true)
	nmat := db.NewAttachedComponent(eGrass, concepts.NamedComponentIndex).(*concepts.Named)
	nmat.Name = "Default Material"
	//tex.Diffuse = color.NRGBA{R: 128, G: 100, B: 50, A: 255}
	tex := materials.ImageFromDb(db, eGrass)
	tex.Source = "data/grass.jpg"
	tex.Filter = false
	tex.GenerateMipMaps = true
	tex.Load()
	//tiled := materials.TiledFromDb(igrass)
	//tiled.Scale = 5.0

	CreateTestGrass(db)
	isky := CreateTestSky(db)
	idirt := CreateTestDirt(db)

	scale := 75
	for x := 0; x < testw; x++ {
		for y := 0; y < testh; y++ {
			sector := CreateTestSector(db, fmt.Sprintf("land_%v_%v", x, y), float64(x*scale), float64(y*scale), float64(scale))
			sector.TopZ.SetAll(300)
			sector.BottomZ.SetAll(rand.Float64() * 30)
			sector.FloorSlope = rand.Float64() * 0.2
			sector.CeilSurface.Material = isky
			for i := 0; i < len(sector.Segments); i++ {
				sector.Segments[i].Surface.Material = isky
				sector.Segments[i].LoSurface.Material = idirt
			}

			if rand.Uint32()%45 == 0 {
				eLight := archetypes.CreateLightBody(db)
				lightBody := core.BodyFromDb(db, eLight)
				lightBody.Pos.Original = concepts.Vector3{float64(x*scale) + rand.Float64()*float64(scale), float64(y*scale) + rand.Float64()*float64(scale), 200}
				lightBody.Pos.ResetToOriginal()
				lightBody.Mass = 0
				log.Println("Generated light")
			}
		}
	}
	for x := 0; x < testw; x++ {
		for y := 0; y < testh; y++ {
			eSector := db.GetEntityByName(fmt.Sprintf("land_%v_%v", x, y))
			sector := core.SectorFromDb(db, eSector)
			// Randomly rotate the segments
			rot := int(rand.Uint32() % 3)
			for r := 0; r < rot; r++ {
				sector.Segments = append(sector.Segments[1:], sector.Segments[0])
			}
		}
	}
	// After everything's loaded, trigger the controllers
	db.ActAllControllers(concepts.ControllerRecalculate)
	AutoPortal(db)
	db.ActAllControllers(concepts.ControllerLoaded)
}
func CreateTestWorld2(db *concepts.EntityComponentDB) {
	eSpawn := archetypes.CreateBasic(db, core.SpawnComponentIndex)
	spawn := core.SpawnFromDb(db, eSpawn)
	spawn.Spawn[0] = 50
	spawn.Spawn[1] = 50
	spawn.Spawn[2] = 50

	CreateTestGrass(db)
	isky := CreateTestSky(db)
	idirt := CreateTestDirt(db)

	sector := CreateTestSector(db, "sector1", -100, -100, 200)
	sector.TopZ.SetAll(100)
	sector.BottomZ.SetAll(0)
	sector2 := CreateTestSector(db, "sector2", 100, -100, 200)
	sector2.TopZ.SetAll(100)
	sector2.BottomZ.SetAll(-10)
	sector3 := CreateTestSector(db, "sector3", 300, -100, 200)
	sector3.TopZ.SetAll(100)
	sector3.BottomZ.SetAll(0)
	sector3.FloorSurface.Material = idirt
	sector3.CeilSurface.Material = isky
	sector3.Segments[1].Surface.Material = isky

	eLight := archetypes.CreateLightBody(db)
	lightBody := core.BodyFromDb(db, eLight)
	lightBody.Pos.Original = concepts.Vector3{0, 0, 60}
	lightBody.Pos.ResetToOriginal()
	lightBody.Mass = 0
	log.Println("Generated light")

	// After everything's loaded, trigger the controllers
	db.ActAllControllers(concepts.ControllerRecalculate)
	AutoPortal(db)
	db.ActAllControllers(concepts.ControllerLoaded)
}
