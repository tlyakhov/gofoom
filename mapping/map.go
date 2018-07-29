package mapping

import (
	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/math"
)

var (
	ValidSectorTypes = map[string]interface{}{
		"Sector": Sector{},
	}
)

type Map struct {
	concepts.Base

	Sectors        []concepts.ISerializable
	Materials      []concepts.ISerializable `editable:"Materials" edit_type:"Material"`
	Player         *Player
	Spawn          *math.Vector3 `editable:"Spawn" edit_type:"Vector"`
	EntitiesPaused bool
}

func (m *Map) ClearLightmaps() {
	for _, item := range m.Sectors {
		if sector, ok := item.(*Sector); ok {
			sector.ClearLightmaps()
		}
	}
}

func (m *Map) Deserialize(data map[string]interface{}) *Map {
	m.Base.Deserialize(data)
	if v, ok := data["EntitiesPaused"]; ok {
		m.EntitiesPaused = v.(bool)
	} else {
		m.EntitiesPaused = false
	}
	if v, ok := data["Sectors"]; ok {
		concepts.DeSeArray(&m.Sectors, v, ValidSectorTypes)
	}
	return m
}
