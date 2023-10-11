package core

import (
	"sync"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/registry"
)

type Map struct {
	concepts.Base

	Sectors        map[string]AbstractSector
	Materials      map[string]concepts.ISerializable `editable:"Materials" edit_type:"Material"`
	Player         AbstractEntity
	Spawn          concepts.Vector3 `editable:"Spawn"`
	EntitiesPaused bool
	RenderLock     sync.Mutex
}

func init() {
	registry.Instance().Register(Map{})
}

func (m *Map) Recalculate() {
	m.RenderLock.Lock()
	defer m.RenderLock.Unlock()
	for _, sector := range m.Sectors {
		sector.Physical().Recalculate()
	}
}

func (m *Map) Initialize() {
	m.Spawn = concepts.Vector3{}
	m.Materials = make(map[string]concepts.ISerializable)
	m.Sectors = make(map[string]AbstractSector)
}

func (m *Map) Deserialize(data map[string]interface{}) {
	m.Initialize()
	m.Base.Deserialize(data)
	if v, ok := data["EntitiesPaused"]; ok {
		m.EntitiesPaused = v.(bool)
	}
	if v, ok := data["SpawnX"]; ok {
		m.Spawn[0] = v.(float64)
	}
	if v, ok := data["SpawnY"]; ok {
		m.Spawn[1] = v.(float64)
	}
	// Load materials first so sectors have access to them.
	if v, ok := data["Materials"]; ok {
		concepts.MapCollection(m, &m.Materials, v)
	}
	if v, ok := data["Sectors"]; ok {
		concepts.MapCollection(m, &m.Sectors, v)
	}
	m.Recalculate()
}

func (m *Map) Serialize() map[string]interface{} {
	result := m.Base.Serialize()
	result["EntitiesPaused"] = m.EntitiesPaused
	result["SpawnX"] = m.Spawn[0]
	result["SpawnY"] = m.Spawn[1]
	materials := []interface{}{}
	for _, mat := range m.Materials {
		materials = append(materials, mat.Serialize())
	}
	result["Materials"] = materials
	sectors := []interface{}{}
	for _, sector := range m.Sectors {
		sectors = append(sectors, sector.Serialize())
	}
	result["Sectors"] = sectors
	return result
}

func (m *Map) DefaultMaterial() concepts.ISerializable {
	if def, ok := m.Materials["Default"]; ok {
		return def
	}

	// Otherwise try a random one?
	for _, mat := range m.Materials {
		return mat
	}
	return nil
}
