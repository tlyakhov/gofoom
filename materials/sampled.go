package materials

import (
	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/registry"
	"github.com/tlyakhov/gofoom/texture"
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
