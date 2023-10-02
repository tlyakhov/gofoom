package materials

import (
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/registry"
)

type Painful struct {
	concepts.Base `editable:"^"`

	Hurt float64 `editable:"Hurt" edit_type:"float"`
}

func init() {
	registry.Instance().Register(Painful{})
}

func (m *Painful) Initialize() {
	m.Base.Initialize()
}

func (m *Painful) Deserialize(data map[string]interface{}) {
	m.Initialize()
	m.Base.Deserialize(data)
	if v, ok := data["Hurt"]; ok {
		m.Hurt = v.(float64)
	}
}

func (m *Painful) Serialize() map[string]interface{} {
	result := m.Base.Serialize()
	result["Type"] = "materials.Painful"
	result["Hurt"] = m.Hurt
	return result
}
