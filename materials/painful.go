package materials

import (
	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/registry"
)

type Painful struct {
	concepts.Base
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
