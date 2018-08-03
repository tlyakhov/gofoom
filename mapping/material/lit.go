package material

import (
	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/registry"
)

type Lit struct {
	*concepts.Base
	Ambient *concepts.Vector3 `editable:"Ambient Color" edit_type:"vector"`
	Diffuse *concepts.Vector3 `editable:"Diffuse Color" edit_type:"vector"`
}

func init() {
	registry.Instance().Register(Lit{})
}

func (m *Lit) Initialize() {
	m.Base = &concepts.Base{}
	m.Base.Initialize()
	m.Ambient = &concepts.Vector3{}
	m.Diffuse = &concepts.Vector3{1, 1, 1}
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
