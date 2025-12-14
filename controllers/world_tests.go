// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"fmt"
	"log"
	"math/rand"
	"tlyakhov/gofoom/archetypes"
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/character"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/inventory"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/ecs"
)

func CreateTestSector(name string, x, y, size float64) *core.Sector {
	eSector := ecs.NewEntity()
	sector := ecs.NewAttachedComponent(eSector, core.SectorCID).(*core.Sector)
	named := ecs.NewAttachedComponent(eSector, ecs.NamedCID).(*ecs.Named)
	named.Name = name

	mat := DefaultMaterial()
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

func CreateTestHeightmapSector(name string, x, y, size float64) (*core.Sector, *core.Sector) {
	eSector := ecs.NewEntity()
	sector1 := ecs.NewAttachedComponent(eSector, core.SectorCID).(*core.Sector)
	named := ecs.NewAttachedComponent(eSector, ecs.NamedCID).(*ecs.Named)
	named.Name = name + "_1"

	sector1.NoShadows = true
	mat := DefaultMaterial()
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

	eSector = ecs.NewEntity()
	sector2 := ecs.NewAttachedComponent(eSector, core.SectorCID).(*core.Sector)
	named = ecs.NewAttachedComponent(eSector, ecs.NamedCID).(*ecs.Named)
	named.Name = name + "_2"

	sector2.NoShadows = true
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

func CreateTestGrass() ecs.Entity {
	eGrass := ecs.NewEntity()
	named := ecs.NewAttachedComponent(eGrass, ecs.NamedCID).(*ecs.Named)
	named.Name = "Default Material"
	//img.Diffuse = color.NRGBA{R: 128, G: 100, B: 50, A: 255}
	img := ecs.NewAttachedComponent(eGrass, materials.ImageCID).(*materials.Image)
	img.Source = "data/textures/grass2.png"
	img.Filter = true
	img.GenerateMipMaps = true
	img.MarkDirty()
	ecs.NewAttachedComponent(eGrass, materials.LitCID)
	return eGrass
}
func CreateTestSky() ecs.Entity {
	skyImage := ecs.NewAttachedComponent(ecs.NewEntity(), materials.ImageCID).(*materials.Image)
	skyImage.Source = "data/textures/Sky.png"
	skyImage.Filter = false
	skyImage.GenerateMipMaps = false
	skyImage.MarkDirty()

	entity := ecs.NewEntity()
	skyShader := ecs.NewAttachedComponent(entity, materials.ShaderCID).(*materials.Shader)
	skyShader.Stages = append(skyShader.Stages, new(materials.ShaderStage))
	skyShader.Stages[0].Construct(nil)
	skyShader.Stages[0].Material = skyImage.Entity
	skyShader.Stages[0].Flags = materials.ShaderSky | materials.ShaderTiled
	named := ecs.NewAttachedComponent(entity, ecs.NamedCID).(*ecs.Named)
	named.Name = "Sky"

	return entity
}

func CreateTestDirt() ecs.Entity {
	eDirt := ecs.NewEntity()
	nmat := ecs.NewAttachedComponent(eDirt, ecs.NamedCID).(*ecs.Named)
	nmat.Name = "Dirt"
	tex := ecs.NewAttachedComponent(eDirt, materials.ImageCID).(*materials.Image)
	tex.Source = "data/textures/FDef.png"
	tex.Filter = false
	tex.GenerateMipMaps = true
	tex.MarkDirty()
	ecs.NewAttachedComponent(eDirt, materials.LitCID)
	return eDirt
}

func CreateTestTree() ecs.Entity {
	eTree := ecs.NewEntity()
	nmat := ecs.NewAttachedComponent(eTree, ecs.NamedCID).(*ecs.Named)
	nmat.Name = "Tree"
	tex := ecs.NewAttachedComponent(eTree, materials.ImageCID).(*materials.Image)
	tex.Source = "data/sprites/tree.png"
	tex.Filter = false
	tex.GenerateMipMaps = true
	tex.MarkDirty()
	ecs.NewAttachedComponent(eTree, materials.LitCID)
	return eTree
}

func CreateSpawn() {
	e := ecs.NewEntity()
	player := &character.Player{}
	player.Construct(nil)
	ecs.AttachTyped(e, &player)
	spawner := &behaviors.Spawner{}
	spawner.Construct(nil)
	ecs.AttachTyped(e, &spawner)
	body := &core.Body{}
	body.Construct(nil)
	body.Pos.SetAll(concepts.Vector3{50, 50, 240})
	ecs.AttachTyped(e, &body)
	mobile := &core.Mobile{}
	mobile.Construct(nil)
	mobile.Mass = 80
	ecs.AttachTyped(e, &mobile)
	alive := &behaviors.Alive{}
	alive.Construct(nil)
	ecs.AttachTyped(e, &alive)
	carrier := &inventory.Carrier{}
	carrier.Construct(nil)
	ecs.AttachTyped(e, &carrier)

	Spawn(spawner)
}

func CreateTestWorld() {
	testw := 30
	testh := 30

	CreateTestGrass()
	eSky := CreateTestSky()
	eDirt := CreateTestDirt()

	scale := 75
	for x := range testw {
		for y := range testh {
			sector := CreateTestSector(fmt.Sprintf("land_%v_%v", x, y), float64(x*scale), float64(y*scale), float64(scale))
			sector.Top.Z.SetAll(300)
			sector.Bottom.Z.SetAll(rand.Float64() * 30)
			//sector.FloorSlope = rand.Float64() * 0.2
			sector.Top.Surface.Material = eSky
			for i := 0; i < len(sector.Segments); i++ {
				sector.Segments[i].Surface.Material = eSky
				sector.Segments[i].LoSurface.Material = eDirt
			}

			if rand.Uint32()%45 == 0 {
				eLight := archetypes.CreateLightBody()
				lightBody := core.GetBody(eLight)
				lightBody.Pos.Spawn = concepts.Vector3{float64(x*scale) + rand.Float64()*float64(scale), float64(y*scale) + rand.Float64()*float64(scale), 200}
				lightBody.Pos.ResetToSpawn()
				log.Println("Generated light")
			}
		}
	}
	for x := range testw {
		for y := range testh {
			eSector := ecs.GetEntityByName(fmt.Sprintf("land_%v_%v", x, y))
			sector := core.GetSector(eSector)
			// Randomly rotate the segments
			rot := int(rand.Uint32() % 3)
			for range rot {
				sector.Segments = append(sector.Segments[1:], sector.Segments[0])
			}
		}
	}
	CreateSpawn()
	// After everything's loaded, trigger the controllers
	ecs.ActAllControllers(ecs.ControllerRecalculate)
	AutoPortal()
}
func CreateTestWorld2() {
	CreateTestGrass()
	isky := CreateTestSky()
	idirt := CreateTestDirt()

	sector := CreateTestSector("sector1", -100, -100, 200)
	sector.Top.Z.SetAll(100)
	sector.Bottom.Z.SetAll(0)
	sector2 := CreateTestSector("sector2", 100, -100, 200)
	sector2.Top.Z.SetAll(100)
	sector2.Bottom.Z.SetAll(-10)
	sector3 := CreateTestSector("sector3", 300, -100, 200)
	sector3.Top.Z.SetAll(100)
	sector3.Bottom.Z.SetAll(0)
	sector3.Bottom.Surface.Material = idirt
	sector3.Top.Surface.Material = isky
	sector3.Segments[1].Surface.Material = isky

	eLight := archetypes.CreateLightBody()
	lightBody := core.GetBody(eLight)
	lightBody.Pos.Spawn = concepts.Vector3{0, 0, 60}
	lightBody.Pos.ResetToSpawn()
	log.Println("Generated light")

	CreateSpawn()

	// After everything's loaded, trigger the controllers
	ecs.ActAllControllers(ecs.ControllerRecalculate)
	AutoPortal()
}

// TODO: Worth adapting this heightmap generator to be usable in editor?
// I think this is probably too performance-intensive for in-game, but maybe
// with small heightmaps it could be viable.
func CreateTestWorld3() {
	// This is a pretty epic stress test: a 64x64 heightmap represents 4096*2 =
	// 8192 triangle-shaped sectors, total 24,576 segments. As of 2024-01-02,
	// the worst performance offender is generating/maintaining the PVS,
	// particularly dynamic recalculation, as well as lighting (which requires
	// traversing the entire map for every lightmap texel!)
	// Without lighting, this actually performs very well - 60fps comfortably.

	heightImage := ecs.NewAttachedComponent(ecs.NewEntity(), materials.ImageCID).(*materials.Image)
	heightImage.Flags |= ecs.EntityInternal
	heightImage.Construct(map[string]any{
		"Source":          "data/textures/test-heightmap.jpg",
		"Filter":          true,
		"GenerateMipMaps": true,
		"ConvertSRGB":     false,
	})
	ecs.ActAllControllersOneEntity(heightImage.Entity, ecs.ControllerRecalculate)

	testw := 64
	testh := 64
	heightmap := make([]float64, testw*testh)
	var sample concepts.Vector4
	for x := range testw {
		for y := range testh {
			//heightmap[y*testw+x] = rand.Float64() * 50
			heightImage.Sample(
				float64(x)/float64(testw),
				float64(y)/float64(testh),
				uint32(testw), uint32(testh), &sample)
			heightmap[y*testw+x] = concepts.RGBtoHSP(sample.To3D())[2] * 250
		}
	}

	eGrass := CreateTestGrass()
	eSky := CreateTestSky()
	eDirt := CreateTestDirt()
	eTree := CreateTestTree()

	scale := 50
	for x := 0; x < testw-1; x++ {
		for y := 0; y < testh-1; y++ {
			sector1, sector2 := CreateTestHeightmapSector(fmt.Sprintf("land_%v_%v", x, y), float64(x*scale), float64(y*scale), float64(scale))
			v1 := &concepts.Vector3{
				sector1.Segments[1].P.Spawn[0] - sector1.Segments[0].P.Spawn[0],
				sector1.Segments[1].P.Spawn[1] - sector1.Segments[0].P.Spawn[1],
				heightmap[(y+0)*testw+x+1] - heightmap[(y+0)*testw+x+0]}
			v2 := &concepts.Vector3{
				sector1.Segments[2].P.Spawn[0] - sector1.Segments[0].P.Spawn[0],
				sector1.Segments[2].P.Spawn[1] - sector1.Segments[0].P.Spawn[1],
				heightmap[(y+1)*testw+x+0] - heightmap[(y+0)*testw+x+0]}
			sector1.Bottom.Normal = *v1.Cross(v2).NormSelf()
			sector1.Bottom.Z.SetAll(heightmap[(y+0)*testw+x+0])
			if sector1.Bottom.Z.Spawn < 20 {
				sector1.Bottom.Surface.Material = eDirt
			} else {
				sector1.Bottom.Surface.Material = eGrass
			}
			v1 = &concepts.Vector3{
				sector2.Segments[1].P.Spawn[0] - sector2.Segments[0].P.Spawn[0],
				sector2.Segments[1].P.Spawn[1] - sector2.Segments[0].P.Spawn[1],
				heightmap[(y+1)*testw+x+1] - heightmap[(y+0)*testw+x+1]}
			v2 = &concepts.Vector3{
				sector2.Segments[2].P.Spawn[0] - sector2.Segments[1].P.Spawn[0],
				sector2.Segments[2].P.Spawn[1] - sector2.Segments[1].P.Spawn[1],
				heightmap[(y+1)*testw+x+0] - heightmap[(y+1)*testw+x+1]}
			sector2.Bottom.Normal = *v1.Cross(v2).NormSelf()
			sector2.Bottom.Z.SetAll(heightmap[(y+0)*testw+x+1])
			if sector2.Bottom.Z.Spawn < 20 {
				sector2.Bottom.Surface.Material = eDirt
			} else {
				sector2.Bottom.Surface.Material = eGrass
			}

			sector1.Top.Z.SetAll(500)
			sector2.Top.Z.SetAll(500)

			sector1.Top.Surface.Material = eSky
			for i := 0; i < len(sector1.Segments); i++ {
				sector1.Segments[i].Surface.Material = eSky
				sector1.Segments[i].LoSurface.Material = eDirt
			}
			sector2.Top.Surface.Material = eSky
			for i := 0; i < len(sector2.Segments); i++ {
				sector2.Segments[i].Surface.Material = eSky
				sector2.Segments[i].LoSurface.Material = eDirt
			}
			sector1.Recalculate()
			sector2.Recalculate()
		}
	}

	for range 8 {
		eLight := archetypes.CreateLightBody()
		lightBody := core.GetBody(eLight)
		lightBody.Pos.Spawn = concepts.Vector3{float64(testw*scale) * rand.Float64(), float64(testh*scale) * rand.Float64(), 450}
		lightBody.Pos.ResetToSpawn()
		light := core.GetLight(eLight)
		light.Strength = 3
		light.Attenuation = 0.3
		bc := BodyController{}
		bc.Target(lightBody, eLight)
		bc.findBodySector()
		log.Println("Generated light")
	}

	for range 32 {
		eTreeBody := ecs.NewEntity()
		body := ecs.NewAttachedComponent(eTreeBody, core.BodyCID).(*core.Body)
		x := float64(testw*scale) * rand.Float64()
		y := float64(testh*scale) * rand.Float64()
		z := heightmap[(int(y/float64(scale))*testw + int(x/float64(scale)))]
		body.Pos.Spawn = concepts.Vector3{x, y, z + 25}
		body.Pos.ResetToSpawn()
		body.Size.Spawn[0] = 50
		body.Size.Spawn[1] = 50
		body.Size.ResetToSpawn()
		vis := ecs.NewAttachedComponent(eTreeBody, materials.VisibleCID).(*materials.Visible)
		vis.Shadow = materials.ShadowImage
		shader := ecs.NewAttachedComponent(eTreeBody, materials.ShaderCID).(*materials.Shader)
		stage := &materials.ShaderStage{}
		stage.Construct(nil)
		stage.Material = eTree
		shader.Stages = append(shader.Stages, stage)
	}
	CreateSpawn()
	AutoPortal()
	// After everything's loaded, trigger the controllers
	ecs.ActAllControllers(ecs.ControllerRecalculate)
}
