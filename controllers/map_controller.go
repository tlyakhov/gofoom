package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"

	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/controllers/provide"
	"tlyakhov/gofoom/core"
	"tlyakhov/gofoom/entities"
	"tlyakhov/gofoom/materials"
	"tlyakhov/gofoom/texture"
)

type MapController struct {
	*core.Map
}

func NewMapController(m *core.Map) *MapController {
	return &MapController{Map: m}
}

func LoadMap(filename string) (*MapController, error) {
	fileContents, err := os.ReadFile(filename)

	if err != nil {
		return nil, err
	}

	var parsed interface{}
	err = json.Unmarshal(fileContents, &parsed)
	if err != nil {
		return nil, err
	}

	m := NewMapController(new(core.Map))
	m.Construct(parsed.(map[string]interface{}))
	m.Player = entities.NewPlayer(m.Map)
	m.Recalculate()
	return m, nil
}

func (m *MapController) Save(filename string) {
	mapped := m.Serialize()
	bytes, err := json.MarshalIndent(mapped, "", "  ")

	if err != nil {
		panic(err)
	}

	os.WriteFile(filename, bytes, os.ModePerm)
}

func (m *MapController) Recalculate() {
	m.Map.Recalculate()
	for _, s := range m.Sectors {
		provide.Passer.For(s).Recalculate()
		for _, e := range s.Physical().Entities {
			if c := provide.Collider.For(e); c != nil {
				c.Collide()
			}
		}
	}
}

func (m *MapController) Frame() {
	player := provide.EntityAnimator.For(m.Player)
	player.Frame()

	for _, sector := range m.Sectors {
		provide.Interactor.For(sector).ActOnEntity(m.Player)
		for _, e := range sector.Physical().Entities {
			if !e.Physical().Active {
				continue
			}
			for _, pvs := range sector.Physical().PVSEntity {
				_ = pvs
				provide.Interactor.For(pvs).ActOnEntity(e)
			}
		}
		provide.SectorAnimator.For(sector).Frame()
	}
}

func (m *MapController) AutoPortal() {
	seen := map[string]bool{}
	for _, sector := range m.Sectors {
		for _, segment := range sector.Physical().Segments {
			segment.AdjacentSector = nil
			segment.AdjacentSegment = nil
			if segment.MidMaterial == nil {
				segment.MidMaterial = m.DefaultMaterial()
			}
		}
	}
	for _, sector := range m.Sectors {
		for _, sector2 := range m.Sectors {
			if sector == sector2 {
				continue
			}
			id := sector.GetBase().ID + "|" + sector2.GetBase().ID
			id2 := sector2.GetBase().ID + "|" + sector.GetBase().ID
			if seen[id2] || seen[id] {
				continue
			}
			seen[id] = true

			//if !sector.Physical().AABBIntersect(&sector2.Physical().Min, &sector2.Physical().Max) {
			//	continue
			//}

			for _, segment := range sector.Physical().Segments {
				for _, segment2 := range sector2.Physical().Segments {
					if segment.Matches(segment2) {
						segment2.AdjacentSector = sector
						segment2.AdjacentSegment = segment
						segment.AdjacentSector = sector2
						segment.AdjacentSegment = segment2
					}
				}

			}
		}
	}
	m.Recalculate()
}

func (ms *MapController) CreateTestSector(id string, x, y, size float64) *core.PhysicalSector {
	mat := ms.Map.Materials["Default"]
	sector := &core.PhysicalSector{}
	sector.Construct(nil)
	sector.GetBase().ID = id
	sector.SetParent(ms.Map)
	ms.Sectors[sector.ID] = sector
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

func (ms *MapController) CreateTest() {
	ms.Player = entities.NewPlayer(ms.Map)
	ms.Spawn[0] = 50
	ms.Spawn[1] = 50
	ms.Spawn[2] = 32
	ms.Player.Physical().Pos.Original = ms.Spawn
	ms.Player.Physical().Pos.Reset()
	mat := &materials.LitSampled{}
	mat.Construct(nil)
	mat.GetBase().ID = "Default"
	//tex := &texture.Solid{Diffuse: color.NRGBA{R: 128, G: 100, B: 50, A: 255}}
	tex := &texture.Image{}
	tex.Construct(nil)
	tex.Source = "data/grass.jpg"
	tex.Filter = true
	tex.GenerateMipMaps = true
	tex.Load()
	mat.Sampler = tex
	mat.Scale = 5.0

	mat.SetParent(ms.Map)
	tex.SetParent(mat)
	ms.Materials[mat.GetBase().ID] = mat
	scale := 75
	for x := 0; x < 20; x++ {
		for y := 0; y < 20; y++ {
			sector := ms.CreateTestSector(fmt.Sprintf("land_%v_%v", x, y), float64(x*scale), float64(y*scale), float64(scale))
			sector.TopZ.Original = 300
			sector.BottomZ.Original = rand.Float64() * 30
			sector.FloorSlope = rand.Float64() * 0.2
			// Randomly rotate the segments
			rot := int(rand.Uint32() % 3)
			for r := 0; r < rot; r++ {
				sector.Segments = append(sector.Segments[1:], sector.Segments[0])
			}

			if rand.Uint32()%40 == 0 {
				light := &entities.Light{}
				light.Construct(nil)
				light.Pos.Original = concepts.Vector3{float64(x*scale) + rand.Float64()*float64(scale), float64(y*scale) + rand.Float64()*float64(scale), 200}
				light.Pos.Reset()
				light.SetParent(sector)
				sector.Entities[light.ID] = light
				log.Println("Generated light")
			}
		}
	}
	sector := ms.Sectors["land_0_0"].(*core.PhysicalSector)

	ms.Player.SetParent(sector)

	ms.Recalculate()
	ms.AutoPortal()
}
