package materials

import (
	"math"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/registry"
	"tlyakhov/gofoom/texture"
)

type Sampled struct {
	concepts.Base `editable:"^"`

	Sampler  texture.ISampler `editable:"Texture" edit_type:"Texture"`
	IsLiquid bool             `editable:"Is Liquid?" edit_type:"bool"`
	Scale    float64          `editable:"Scale" edit_type:"float"`
}

func init() {
	registry.Instance().Register(Sampled{})
}

func (m *Sampled) Construct(data map[string]interface{}) {
	m.Base.Construct(data)
	m.Model = m
	m.Scale = 1.0

	if data == nil {
		return
	}

	if v, ok := data["IsLiquid"]; ok {
		m.IsLiquid = v.(bool)
	}
	if v, ok := data["Texture"]; ok {
		m.Sampler = concepts.MapPolyStruct(m, v.(map[string]interface{})).(texture.ISampler)
	}
	if v, ok := data["Scale"]; ok {
		m.Scale = v.(float64)
	}

}

func (m *Sampled) Serialize() map[string]interface{} {
	result := m.Base.Serialize()
	result["Type"] = "materials.Sampled"
	result["IsLiquid"] = m.IsLiquid
	result["Scale"] = m.Scale
	result["Texture"] = m.Sampler.Serialize()
	return result
}

func (m *Sampled) Sample(u, v float64, light *concepts.Vector3, scale float64) uint32 {
	u -= math.Floor(u)
	v -= math.Floor(v)
	u = math.Abs(u)
	v = math.Abs(v)

	return m.Sampler.Sample(u/m.Scale, v/m.Scale, scale/m.Scale)
}
