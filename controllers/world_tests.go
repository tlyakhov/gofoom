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
	isector := archetypes.CreateSector(db)
	sector := core.SectorFromDb(isector)
	named := db.NewComponent(isector.Entity, concepts.NamedComponentIndex).(*concepts.Named)
	named.Name = name

	mat := DefaultMaterial(db)
	sector.FloorMaterial = mat
	sector.CeilMaterial = mat
	seg := sector.AddSegment(x, y)
	seg.MidMaterial = mat
	seg.HiMaterial = mat
	seg.LoMaterial = mat
	seg = sector.AddSegment(x+size, y)
	seg.MidMaterial = mat
	seg.HiMaterial = mat
	seg.LoMaterial = mat
	seg = sector.AddSegment(x+size, y+size)
	seg.MidMaterial = mat
	seg.HiMaterial = mat
	seg.LoMaterial = mat
	seg = sector.AddSegment(x, y+size)
	seg.MidMaterial = mat
	seg.HiMaterial = mat
	seg.LoMaterial = mat

	sector.Recalculate()
	return sector
}

func CreateTestGrass(db *concepts.EntityComponentDB) *concepts.EntityRef {
	igrass := archetypes.CreateBasicMaterial(db, true)
	nmat := db.NewComponent(igrass.Entity, concepts.NamedComponentIndex).(*concepts.Named)
	nmat.Name = "Default Material"
	//tex.Diffuse = color.NRGBA{R: 128, G: 100, B: 50, A: 255}
	tex := materials.ImageFromDb(igrass)
	tex.Source = "data/grass.jpg"
	tex.Filter = true
	tex.GenerateMipMaps = true
	tex.Load()
	//tiled := materials.TiledFromDb(igrass)
	//tiled.Scale = 5.0
	return igrass
}
func CreateTestSky(db *concepts.EntityComponentDB) *concepts.EntityRef {
	isky := archetypes.CreateBasic(db, materials.SkyComponentIndex)
	nmat := db.NewComponent(isky.Entity, concepts.NamedComponentIndex).(*concepts.Named)
	nmat.Name = "Sky"
	tex := db.NewComponent(isky.Entity, materials.ImageComponentIndex).(*materials.Image)
	tex.Source = "data/Sky.png"
	tex.Filter = false
	tex.GenerateMipMaps = false
	tex.Load()
	return isky
}

func CreateTestDirt(db *concepts.EntityComponentDB) *concepts.EntityRef {
	idirt := archetypes.CreateBasicMaterial(db, true)
	nmat := db.NewComponent(idirt.Entity, concepts.NamedComponentIndex).(*concepts.Named)
	nmat.Name = "Dirt"
	tex := materials.ImageFromDb(idirt)
	tex.Source = "data/FDef.png"
	tex.Filter = false
	tex.GenerateMipMaps = true
	tex.Load()
	materials.TiledFromDb(idirt)
	return idirt
}

func CreateTestWorld(db *concepts.EntityComponentDB) {
	testw := 30
	testh := 30
	ispawn := archetypes.CreateBasic(db, core.SpawnComponentIndex)
	spawn := core.SpawnFromDb(ispawn)
	spawn.Spawn[0] = 250
	spawn.Spawn[1] = 250
	spawn.Spawn[2] = 100

	igrass := archetypes.CreateBasicMaterial(db, true)
	nmat := db.NewComponent(igrass.Entity, concepts.NamedComponentIndex).(*concepts.Named)
	nmat.Name = "Default Material"
	//tex.Diffuse = color.NRGBA{R: 128, G: 100, B: 50, A: 255}
	tex := materials.ImageFromDb(igrass)
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
			sector.TopZ.Set(300)
			sector.BottomZ.Set(rand.Float64() * 30)
			sector.FloorSlope = rand.Float64() * 0.2
			sector.CeilMaterial = isky
			for i := 0; i < len(sector.Segments); i++ {
				sector.Segments[i].MidMaterial = isky
				sector.Segments[i].LoMaterial = idirt
			}

			if rand.Uint32()%45 == 0 {
				ilight := archetypes.CreateLightBody(db)
				lightBody := core.BodyFromDb(ilight)
				lightBody.Pos.Original = concepts.Vector3{float64(x*scale) + rand.Float64()*float64(scale), float64(y*scale) + rand.Float64()*float64(scale), 200}
				lightBody.Pos.Reset()
				lightBody.Mass = 0
				log.Println("Generated light")
			}
		}
	}
	for x := 0; x < testw; x++ {
		for y := 0; y < testh; y++ {
			isector := db.GetEntityRefByName(fmt.Sprintf("land_%v_%v", x, y))
			sector := core.SectorFromDb(isector)
			// Randomly rotate the segments
			rot := int(rand.Uint32() % 3)
			for r := 0; r < rot; r++ {
				sector.Segments = append(sector.Segments[1:], sector.Segments[0])
			}
		}
	}
	// After everything's loaded, trigger the controllers
	set := db.NewControllerSet()
	set.ActGlobal(concepts.ControllerRecalculate)
	AutoPortal(db)
	set.ActGlobal(concepts.ControllerLoaded)
}
func CreateTestWorld2(db *concepts.EntityComponentDB) {
	ispawn := archetypes.CreateBasic(db, core.SpawnComponentIndex)
	spawn := core.SpawnFromDb(ispawn)
	spawn.Spawn[0] = 50
	spawn.Spawn[1] = 50
	spawn.Spawn[2] = 50

	CreateTestGrass(db)
	isky := CreateTestSky(db)
	idirt := CreateTestDirt(db)

	sector := CreateTestSector(db, "sector1", -100, -100, 200)
	sector.TopZ.Set(100)
	sector.BottomZ.Set(0)
	sector2 := CreateTestSector(db, "sector2", 100, -100, 200)
	sector2.TopZ.Set(100)
	sector2.BottomZ.Set(-10)
	sector3 := CreateTestSector(db, "sector3", 300, -100, 200)
	sector3.TopZ.Set(100)
	sector3.BottomZ.Set(0)
	sector3.FloorMaterial = idirt
	sector3.CeilMaterial = isky
	sector3.Segments[1].MidMaterial = isky

	ilight := archetypes.CreateLightBody(db)
	lightBody := core.BodyFromDb(ilight)
	lightBody.Pos.Original = concepts.Vector3{0, 0, 60}
	lightBody.Pos.Reset()
	lightBody.Mass = 0
	log.Println("Generated light")

	// After everything's loaded, trigger the controllers
	set := db.NewControllerSet()
	set.ActGlobal(concepts.ControllerRecalculate)
	AutoPortal(db)
	set.ActGlobal(concepts.ControllerLoaded)
}
