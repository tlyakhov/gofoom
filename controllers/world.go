package controllers

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"

	"tlyakhov/gofoom/archetypes"
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"
)

type WorldController struct {
	concepts.BaseController
	*core.Spawn
}

func init() {
	concepts.DbTypes().RegisterController(WorldController{})
}

func (m *WorldController) Target(target *concepts.EntityRef) bool {
	m.TargetEntity = target
	m.Spawn = core.SpawnFromDb(target)
	return m.Spawn != nil && m.Active
}

func (a *WorldController) Loaded() {
	// Create a player if we don't have one
	if a.DB.First(behaviors.PlayerComponentIndex) == nil {
		player := archetypes.CreatePlayerMob(a.DB)
		playerMob := core.MobFromDb(player)
		playerMob.Pos.Original = a.Spawn.Spawn
		playerMob.Pos.Reset()
	}
}

func (m *WorldController) Always() {
	for entity, c := range m.DB.Components[core.MobComponentIndex] {
		mob := c.(*core.Mob)
		if !mob.Active {
			continue
		}
		for _, pvs := range mob.Sector().PVSMob {
			m.ControllerSet.Act(m.DB.EntityRef(entity), pvs.EntityRef(), "Proximity")
		}
	}
}

func AutoPortal(db *concepts.EntityComponentDB) {
	seen := map[string]bool{}
	for _, c := range db.All(core.SectorComponentIndex) {
		sector := c.(*core.Sector)
		for _, segment := range sector.Segments {
			segment.AdjacentSector.Reset()
			segment.AdjacentSegment = nil
			if segment.MidMaterial.Nil() {
				segment.MidMaterial = DefaultMaterial(db)
			}
		}
	}
	for _, c := range db.All(core.SectorComponentIndex) {
		for _, c2 := range db.All(core.SectorComponentIndex) {
			if c == c2 {
				continue
			}
			name := strconv.FormatUint(c.GetEntity(), 10) + "|" + strconv.FormatUint(c2.GetEntity(), 10)
			id2 := strconv.FormatUint(c2.GetEntity(), 10) + "|" + strconv.FormatUint(c.GetEntity(), 10)
			if seen[id2] || seen[name] {
				continue
			}
			seen[name] = true

			sector := c.(*core.Sector)
			sector2 := c2.(*core.Sector)
			//if !sector.AABBIntersect(&sector2.Min, &sector2.Max) {
			//	continue
			//}

			for _, segment := range sector.Segments {
				for _, segment2 := range sector2.Segments {
					if segment.Matches(segment2) {
						segment2.AdjacentSector = sector.EntityRef()
						segment2.AdjacentSegment = segment
						segment.AdjacentSector = sector2.EntityRef()
						segment.AdjacentSegment = segment2
					}
				}

			}
		}
	}
}

func DefaultMaterial(db *concepts.EntityComponentDB) *concepts.EntityRef {
	er := db.GetEntityRefByName("Default Material")
	if !er.Nil() {
		return er
	}

	// Otherwise try a random one?
	return db.First(materials.LitComponentIndex).EntityRef()
}

func CreateTestSector(db *concepts.EntityComponentDB, name string, x, y, size float64) *core.Sector {
	er := archetypes.CreateSector(db)
	sector := core.SectorFromDb(er)

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

	return sector
}

func CreateTestWorld(db *concepts.EntityComponentDB) {
	spawner := archetypes.CreateBasic(db, core.SpawnComponentIndex)
	spawn := core.SpawnFromDb(spawner)
	spawn.Spawn[0] = 50
	spawn.Spawn[1] = 50
	spawn.Spawn[2] = 32

	materialer := archetypes.CreateBasicMaterial(db, true)
	nmat := concepts.NamedFromDb(materialer)
	nmat.Name = "Default Material"
	//tex.Diffuse = color.NRGBA{R: 128, G: 100, B: 50, A: 255}
	tex := materials.ImageFromDb(materialer)
	tex.Source = "data/grass.jpg"
	tex.Filter = true
	tex.GenerateMipMaps = true
	tex.Load()
	tiled := materials.TiledFromDb(materialer)
	tiled.Scale = 5.0

	scale := 75
	for x := 0; x < 20; x++ {
		for y := 0; y < 20; y++ {
			sector := CreateTestSector(db, fmt.Sprintf("land_%v_%v", x, y), float64(x*scale), float64(y*scale), float64(scale))
			sector.TopZ.Original = 300
			sector.BottomZ.Original = rand.Float64() * 30
			sector.FloorSlope = rand.Float64() * 0.2
			// Randomly rotate the segments
			rot := int(rand.Uint32() % 3)
			for r := 0; r < rot; r++ {
				sector.Segments = append(sector.Segments[1:], sector.Segments[0])
			}

			if rand.Uint32()%40 == 0 {
				lighter := archetypes.CreateLightMob(db)
				lightMob := core.MobFromDb(lighter)
				lightMob.Pos.Original = concepts.Vector3{float64(x*scale) + rand.Float64()*float64(scale), float64(y*scale) + rand.Float64()*float64(scale), 200}
				lightMob.Pos.Reset()
				log.Println("Generated light")
			}
		}
	}
	// After everything's loaded, trigger the controllers
	set := db.NewControllerSet()
	set.ActGlobal("Loaded")
	set.ActGlobal("Recalculate")
}
