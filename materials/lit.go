package materials

import (
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/registry"
)

type Lit struct {
	concepts.Base `editable:"^"`

	Ambient concepts.Vector3 `editable:"Ambient Color" edit_type:"color"`
	Diffuse concepts.Vector3 `editable:"Diffuse Color" edit_type:"color"`
}

func init() {
	registry.Instance().Register(Lit{})
}

func (m *Lit) Initialize() {
	m.Base.Initialize()
	m.Ambient = concepts.Vector3{0.1, 0.1, 0.1}
	m.Diffuse = concepts.Vector3{1, 1, 1}
}

func (m *Lit) Deserialize(data map[string]interface{}) {
	m.Initialize()
	m.Base.Deserialize(data)
	if v, ok := data["Ambient"]; ok {
		m.Ambient.Deserialize(v.(map[string]interface{}))
	}
	if v, ok := data["Diffuse"]; ok {
		m.Diffuse.Deserialize(v.(map[string]interface{}))
	}
}

func (m *Lit) Serialize() map[string]interface{} {
	result := m.Base.Serialize()
	result["Type"] = "materials.Lit"
	result["Ambient"] = m.Ambient.Serialize()
	result["Diffuse"] = m.Diffuse.Serialize()
	return result
}

func (m *Lit) Sample(u, v float64, light *concepts.Vector3, scale float64) uint32 {
	sum := m.Diffuse.Mul3(light.Add(&m.Ambient)).Mul(255.0).Clamp(0.0, 255.0)
	return sum.ToInt32Color()
}
