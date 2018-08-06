package mapping

import (
	"fmt"

	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/registry"
)

type Map struct {
	concepts.Base

	Sectors        map[string]AbstractSector
	Materials      map[string]concepts.ISerializable `editable:"Materials" edit_type:"Material"`
	Player         *Player
	Spawn          *concepts.Vector3 `editable:"Spawn" edit_type:"Vector"`
	EntitiesPaused bool
}

func init() {
	registry.Instance().Register(Map{})
}

func (m *Map) Recalculate() {
	for _, item := range m.Sectors {
		if sector, ok := item.(*Sector); ok {
			sector.Recalculate()
		}
	}
}

func (m *Map) ClearLightmaps() {
	for _, item := range m.Sectors {
		if sector, ok := item.(*Sector); ok {
			sector.ClearLightmaps()
		}
	}
}

func (m *Map) Initialize() {
	m.Spawn = &concepts.Vector3{}
	m.Materials = make(map[string]concepts.ISerializable)
	m.Sectors = make(map[string]AbstractSector)
	m.Player = &Player{}
	m.Player.Initialize()
	m.Player.Map = m
}

func (m *Map) Deserialize(data map[string]interface{}) {
	m.Initialize()
	m.Base.Deserialize(data)
	if v, ok := data["EntitiesPaused"]; ok {
		m.EntitiesPaused = v.(bool)
	}
	if v, ok := data["SpawnX"]; ok {
		m.Spawn.X = v.(float64)
		m.Player.Pos.X = m.Spawn.X
	}
	if v, ok := data["SpawnY"]; ok {
		m.Spawn.Y = v.(float64)
		m.Player.Pos.Y = m.Spawn.Y
	}
	// Load materials first so sectors have access to them.
	if v, ok := data["Materials"]; ok {
		concepts.MapCollection(m, &m.Materials, v)
		fmt.Printf("Materials: %v\n", m.Materials)
	}
	if v, ok := data["Sectors"]; ok {
		concepts.MapCollection(m, &m.Sectors, v)
	}
	m.Recalculate()
}
