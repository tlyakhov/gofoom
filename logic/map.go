package logic

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/tlyakhov/gofoom/core"
	"github.com/tlyakhov/gofoom/entities"
	"github.com/tlyakhov/gofoom/logic/provide"
)

type MapService struct {
	*core.Map
}

func NewMapService(m *core.Map) *MapService {
	return &MapService{Map: m}
}

func LoadMap(filename string) *MapService {
	fileContents, err := ioutil.ReadFile(filename)

	if err != nil {
		panic(err)
	}
	var parsed interface{}
	err = json.Unmarshal(fileContents, &parsed)
	m := NewMapService(&core.Map{})
	m.Initialize()
	m.Deserialize(parsed.(map[string]interface{}))
	m.Player = entities.NewPlayer(m.Map)
	m.Recalculate()
	return m
}

func (m *MapService) Save(filename string) {
	mapped := m.Serialize()
	bytes, err := json.MarshalIndent(mapped, "", "  ")

	if err != nil {
		panic(err)
	}

	ioutil.WriteFile(filename, bytes, os.ModePerm)
}

func (m *MapService) Recalculate() {
	m.Map.Recalculate()
	for _, s := range m.Sectors {
		provide.Passer.For(s).Recalculate()
		for _, e := range s.Physical().Entities {
			if c, ok := provide.Collider.For(e); ok {
				c.Collide()
			}
		}
	}
}

func (m *MapService) Frame(lastFrameTime float64) {
	player := provide.EntityAnimator.For(m.Player)
	player.Frame(lastFrameTime)

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
		provide.SectorAnimator.For(sector).Frame(lastFrameTime)
	}
}
