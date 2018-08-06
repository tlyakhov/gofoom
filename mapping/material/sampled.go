package material

import (
	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/registry"
	"github.com/tlyakhov/gofoom/texture"
)

type Sampled struct {
	concepts.Base
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
