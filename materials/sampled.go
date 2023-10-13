package materials

import (
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/registry"
	"tlyakhov/gofoom/texture"
)

type Sampled struct {
	concepts.Base `editable:"^"`

	Sampler  texture.ISampler `editable:"Texture" edit_type:"Texture"`
	IsLiquid bool             `editable:"Is Liquid?" edit_type:"bool"`
}

func init() {
	registry.Instance().Register(Sampled{})
}

func (m *Sampled) Initialize() {
	m.Base.Initialize()
}

func (m *Sampled) Deserialize(data map[string]interface{}) {
	m.Initialize()
	m.Base.Deserialize(data)
	if v, ok := data["IsLiquid"]; ok {
		m.IsLiquid = v.(bool)
	}
	if v, ok := data["Texture"]; ok {
		m.Sampler = concepts.MapPolyStruct(m, v.(map[string]interface{})).(texture.ISampler)
	}
}

func (m *Sampled) Serialize() map[string]interface{} {
	result := m.Base.Serialize()
	result["Type"] = "materials.Sampled"
	result["IsLiquid"] = m.IsLiquid
	result["Texture"] = m.Sampler.Serialize()
	return result
}

func (m *Sampled) Sample(u, v float64, light *concepts.Vector3, scale float64) uint32 {
	for ; u < 0; u++ {
	}
	for ; u > 1; u-- {
	}
	for ; v < 0; v++ {
	}
	for ; v > 1; v-- {
	}

	return m.Sampler.Sample(u, v, scale)
}
