package core

import (
	"sync"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/registry"
)

type Map struct {
	concepts.Base

	Simulation *Simulation
	Sectors    map[string]AbstractSector
	Materials  map[string]Sampleable `editable:"Materials" edit_type:"Material"`
	Player     AbstractMob
	Spawn      concepts.Vector3 `editable:"Spawn"`
	MobsPaused bool
	RenderLock sync.Mutex
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

func (m *Map) Attach(sim *Simulation) {
	m.Simulation = sim
	m.Player.Physical().Attach(sim)
	for _, s := range m.Sectors {
		if simmed, ok := s.(Simulated); ok {
			simmed.Attach(sim)
		}
	}
	for _, m := range m.Materials {
		if simmed, ok := m.(Simulated); ok {
			simmed.Attach(sim)
		}
	}
}
func (m *Map) Detach() {
	if m.Simulation == nil {
		return
	}
	m.Player.Physical().Detach()
	for _, s := range m.Sectors {
		if simmed, ok := s.(Simulated); ok {
			simmed.Detach()
		}
	}
	for _, m := range m.Materials {
		if simmed, ok := m.(Simulated); ok {
			simmed.Detach()
		}
	}
	m.Simulation = nil
}

func (m *Map) Sim() *Simulation {
	return m.Simulation
}

func (m *Map) Construct(data map[string]interface{}) {
	m.Base.Construct(data)
	m.Model = m
	m.Spawn = concepts.Vector3{}
	m.Materials = make(map[string]Sampleable)
	m.Sectors = make(map[string]AbstractSector)

	if data == nil {
		return
	}

	if v, ok := data["MobsPaused"]; ok {
		m.MobsPaused = v.(bool)
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
	if m.Sim() != nil {
		m.Attach(m.Sim())
	}
	m.Recalculate()
}

func (m *Map) Serialize() map[string]interface{} {
	result := m.Base.Serialize()
	result["MobsPaused"] = m.MobsPaused
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

func (m *Map) DefaultMaterial() Sampleable {
	if def, ok := m.Materials["Default"]; ok {
		return def
	}

	// Otherwise try a random one?
	for _, mat := range m.Materials {
		return mat
	}
	return nil
}
