// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"fmt"
	"log"
	"math/rand"
	"tlyakhov/gofoom/archetypes"
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/ecs"
)

func CreateTestSector(db *ecs.ECS, name string, x, y, size float64) *core.Sector {
	eSector := db.NewEntity()
	sector := db.NewAttachedComponent(eSector, core.SectorCID).(*core.Sector)
	named := db.NewAttachedComponent(eSector, ecs.NamedCID).(*ecs.Named)
	named.Name = name

	mat := DefaultMaterial(db)
	sector.Bottom.Surface.Material = mat
	sector.Top.Surface.Material = mat
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

func CreateTestHeightmapSector(db *ecs.ECS, name string, x, y, size float64) (*core.Sector, *core.Sector) {
	eSector := db.NewEntity()
	sector1 := db.NewAttachedComponent(eSector, core.SectorCID).(*core.Sector)
	named := db.NewAttachedComponent(eSector, ecs.NamedCID).(*ecs.Named)
	named.Name = name + "_1"

	mat := DefaultMaterial(db)
	sector1.Bottom.Surface.Material = mat
	sector1.Top.Surface.Material = mat
	seg := sector1.AddSegment(x, y)
	seg.Surface.Material = mat
	seg.HiSurface.Material = mat
	seg.LoSurface.Material = mat
	seg = sector1.AddSegment(x+size, y)
	seg.Surface.Material = mat
	seg.HiSurface.Material = mat
	seg.LoSurface.Material = mat
	seg = sector1.AddSegment(x, y+size)
	seg.Surface.Material = mat
	seg.HiSurface.Material = mat
	seg.LoSurface.Material = mat

	eSector = db.NewEntity()
	sector2 := db.NewAttachedComponent(eSector, core.SectorCID).(*core.Sector)
	named = db.NewAttachedComponent(eSector, ecs.NamedCID).(*ecs.Named)
	named.Name = name + "_2"

	sector2.Bottom.Surface.Material = mat
	sector2.Top.Surface.Material = mat
	seg = sector2.AddSegment(x+size, y)
	seg.Surface.Material = mat
	seg.HiSurface.Material = mat
	seg.LoSurface.Material = mat
	seg = sector2.AddSegment(x+size, y+size)
	seg.Surface.Material = mat
	seg.HiSurface.Material = mat
	seg.LoSurface.Material = mat
	seg = sector2.AddSegment(x, y+size)
	seg.Surface.Material = mat
	seg.HiSurface.Material = mat
	seg.LoSurface.Material = mat

	return sector1, sector2
}

func CreateTestGrass(db *ecs.ECS) ecs.Entity {
	eGrass := archetypes.CreateBasicMaterial(db, true)
	nmat := db.NewAttachedComponent(eGrass, ecs.NamedCID).(*ecs.Named)
	nmat.Name = "Default Material"
	//tex.Diffuse = color.NRGBA{R: 128, G: 100, B: 50, A: 255}
	tex := materials.GetImage(db, eGrass)
	tex.Source = "data/grass.jpg"
	tex.Filter = true
	tex.GenerateMipMaps = true
	tex.Load()
	return eGrass
}
func CreateTestSky(db *ecs.ECS) ecs.Entity {
	img := db.NewAttachedComponent(db.NewEntity(), materials.ImageCID).(*materials.Image)
	img.Source = "data/Sky.png"
	img.Filter = false
	img.GenerateMipMaps = false
	img.Load()

	entity := db.NewEntity()
	sky := db.NewAttachedComponent(entity, materials.ShaderCID).(*materials.Shader)
	sky.Stages = append(sky.Stages, new(materials.ShaderStage))
	sky.Stages[0].Construct(nil)
	sky.Stages[0].Material = img.Entity
	sky.Stages[0].Flags = materials.ShaderSky | materials.ShaderTiled
	named := db.NewAttachedComponent(entity, ecs.NamedCID).(*ecs.Named)
	named.Name = "Sky"

	return entity
}

func CreateTestDirt(db *ecs.ECS) ecs.Entity {
	eDirt := archetypes.CreateBasicMaterial(db, true)
	nmat := db.NewAttachedComponent(eDirt, ecs.NamedCID).(*ecs.Named)
	nmat.Name = "Dirt"
	tex := materials.GetImage(db, eDirt)
	tex.Source = "data/FDef.png"
	tex.Filter = false
	tex.GenerateMipMaps = true
	tex.Load()
	return eDirt
}

func CreateSpawn(db *ecs.ECS) {
	e := db.NewEntity()
	player := &behaviors.Player{}
	player.Construct(nil)
	player.Spawn = true
	db.Upsert(behaviors.PlayerCID, e, player)
	body := &core.Body{}
	body.Construct(nil)
	body.Pos.SetAll(concepts.Vector3{50, 50, 40})
	db.Upsert(core.BodyCID, e, body)
	mobile := &core.Mobile{}
	mobile.Construct(nil)
	mobile.Mass = 80
	db.Upsert(core.MobileCID, e, mobile)
	alive := &behaviors.Alive{}
	alive.Construct(nil)
	db.Upsert(behaviors.AliveCID, e, alive)
	Respawn(db, true)
}

func CreateTestWorld(db *ecs.ECS) {
	testw := 30
	testh := 30

	eGrass := archetypes.CreateBasicMaterial(db, true)
	nmat := db.NewAttachedComponent(eGrass, ecs.NamedCID).(*ecs.Named)
	nmat.Name = "Default Material"
	//tex.Diffuse = color.NRGBA{R: 128, G: 100, B: 50, A: 255}
	tex := materials.GetImage(db, eGrass)
	tex.Source = "data/grass.jpg"
	tex.Filter = false
	tex.GenerateMipMaps = true
	tex.Load()
	//tiled := materials.GetTiled(igrass)
	//tiled.Scale = 5.0

	CreateTestGrass(db)
	isky := CreateTestSky(db)
	idirt := CreateTestDirt(db)

	scale := 75
	for x := 0; x < testw; x++ {
		for y := 0; y < testh; y++ {
			sector := CreateTestSector(db, fmt.Sprintf("land_%v_%v", x, y), float64(x*scale), float64(y*scale), float64(scale))
			sector.Top.Z.SetAll(300)
			sector.Bottom.Z.SetAll(rand.Float64() * 30)
			//sector.FloorSlope = rand.Float64() * 0.2
			sector.Top.Surface.Material = isky
			for i := 0; i < len(sector.Segments); i++ {
				sector.Segments[i].Surface.Material = isky
				sector.Segments[i].LoSurface.Material = idirt
			}

			if rand.Uint32()%45 == 0 {
				eLight := archetypes.CreateLightBody(db)
				lightBody := core.GetBody(db, eLight)
				lightBody.Pos.Original = concepts.Vector3{float64(x*scale) + rand.Float64()*float64(scale), float64(y*scale) + rand.Float64()*float64(scale), 200}
				lightBody.Pos.ResetToOriginal()
				log.Println("Generated light")
			}
		}
	}
	for x := 0; x < testw; x++ {
		for y := 0; y < testh; y++ {
			eSector := db.GetEntityByName(fmt.Sprintf("land_%v_%v", x, y))
			sector := core.GetSector(db, eSector)
			// Randomly rotate the segments
			rot := int(rand.Uint32() % 3)
			for r := 0; r < rot; r++ {
				sector.Segments = append(sector.Segments[1:], sector.Segments[0])
			}
		}
	}
	CreateSpawn(db)
	// After everything's loaded, trigger the controllers
	db.ActAllControllers(ecs.ControllerRecalculate)
	AutoPortal(db)
	db.ActAllControllers(ecs.ControllerLoaded)
}
func CreateTestWorld2(db *ecs.ECS) {
	CreateTestGrass(db)
	isky := CreateTestSky(db)
	idirt := CreateTestDirt(db)

	sector := CreateTestSector(db, "sector1", -100, -100, 200)
	sector.Top.Z.SetAll(100)
	sector.Bottom.Z.SetAll(0)
	sector2 := CreateTestSector(db, "sector2", 100, -100, 200)
	sector2.Top.Z.SetAll(100)
	sector2.Bottom.Z.SetAll(-10)
	sector3 := CreateTestSector(db, "sector3", 300, -100, 200)
	sector3.Top.Z.SetAll(100)
	sector3.Bottom.Z.SetAll(0)
	sector3.Bottom.Surface.Material = idirt
	sector3.Top.Surface.Material = isky
	sector3.Segments[1].Surface.Material = isky

	eLight := archetypes.CreateLightBody(db)
	lightBody := core.GetBody(db, eLight)
	lightBody.Pos.Original = concepts.Vector3{0, 0, 60}
	lightBody.Pos.ResetToOriginal()
	log.Println("Generated light")

	CreateSpawn(db)

	// After everything's loaded, trigger the controllers
	db.ActAllControllers(ecs.ControllerRecalculate)
	AutoPortal(db)
	db.ActAllControllers(ecs.ControllerLoaded)
}

// TODO: Worth adapting this heightmap generator to be usable in editor?
// I think this is probably too performance-intensive for in-game, but maybe
// with small heightmaps it could be viable.
func CreateTestWorld3(db *ecs.ECS) {
	// This is a pretty epic stress test: a 30*30 heightmap represents 900*2 =
	// 1800 triangle-shaped sectors, total 5400 segments.
	testw := 30
	testh := 30
	heightmap := make([]float64, testw*testh)
	for x := 0; x < testw; x++ {
		for y := 0; y < testh; y++ {
			heightmap[y*testw+x] = rand.Float64() * 50
		}
	}

	eGrass := archetypes.CreateBasicMaterial(db, true)
	nmat := db.NewAttachedComponent(eGrass, ecs.NamedCID).(*ecs.Named)
	nmat.Name = "Default Material"
	//tex.Diffuse = color.NRGBA{R: 128, G: 100, B: 50, A: 255}
	tex := materials.GetImage(db, eGrass)
	tex.Source = "data/grass.jpg"
	tex.Filter = false
	tex.GenerateMipMaps = true
	tex.Load()
	//tiled := materials.GetTiled(igrass)
	//tiled.Scale = 5.0

	CreateTestGrass(db)
	isky := CreateTestSky(db)
	idirt := CreateTestDirt(db)

	scale := 75
	for x := 0; x < testw-1; x++ {
		for y := 0; y < testh-1; y++ {
			sector1, sector2 := CreateTestHeightmapSector(db, fmt.Sprintf("land_%v_%v", x, y), float64(x*scale), float64(y*scale), float64(scale))
			v1 := &concepts.Vector3{
				sector1.Segments[1].P[0] - sector1.Segments[0].P[0],
				sector1.Segments[1].P[1] - sector1.Segments[0].P[1],
				heightmap[(y+0)*testw+x+1] - heightmap[(y+0)*testw+x+0]}
			v2 := &concepts.Vector3{
				sector1.Segments[2].P[0] - sector1.Segments[0].P[0],
				sector1.Segments[2].P[1] - sector1.Segments[0].P[1],
				heightmap[(y+1)*testw+x+0] - heightmap[(y+0)*testw+x+0]}
			sector1.Bottom.Normal = *v1.Cross(v2).NormSelf()
			sector1.Bottom.Z.SetAll(heightmap[(y+0)*testw+x+0])
			v1 = &concepts.Vector3{
				sector2.Segments[1].P[0] - sector2.Segments[0].P[0],
				sector2.Segments[1].P[1] - sector2.Segments[0].P[1],
				heightmap[(y+1)*testw+x+1] - heightmap[(y+0)*testw+x+1]}
			v2 = &concepts.Vector3{
				sector2.Segments[2].P[0] - sector2.Segments[1].P[0],
				sector2.Segments[2].P[1] - sector2.Segments[1].P[1],
				heightmap[(y+1)*testw+x+0] - heightmap[(y+1)*testw+x+1]}
			sector2.Bottom.Normal = *v1.Cross(v2).NormSelf()
			sector2.Bottom.Z.SetAll(heightmap[(y+0)*testw+x+1])

			sector1.Top.Z.SetAll(300)
			sector2.Top.Z.SetAll(300)

			sector1.Top.Surface.Material = isky
			for i := 0; i < len(sector1.Segments); i++ {
				sector1.Segments[i].Surface.Material = isky
				sector1.Segments[i].LoSurface.Material = idirt
			}
			sector2.Top.Surface.Material = isky
			for i := 0; i < len(sector2.Segments); i++ {
				sector2.Segments[i].Surface.Material = isky
				sector2.Segments[i].LoSurface.Material = idirt
			}
			sector1.Recalculate()
			sector2.Recalculate()

			if rand.Uint32()%45 == 0 {
				eLight := archetypes.CreateLightBody(db)
				lightBody := core.GetBody(db, eLight)
				lightBody.Pos.Original = concepts.Vector3{float64(x*scale) + rand.Float64()*float64(scale), float64(y*scale) + rand.Float64()*float64(scale), 200}
				lightBody.Pos.ResetToOriginal()
				log.Println("Generated light")
			}
		}
	}

	CreateSpawn(db)
	// After everything's loaded, trigger the controllers
	db.ActAllControllers(ecs.ControllerRecalculate)
	AutoPortal(db)
	db.ActAllControllers(ecs.ControllerLoaded)
}
