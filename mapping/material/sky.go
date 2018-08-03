package material

import (
	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/registry"
)

type Sky struct {
	*concepts.Base
	Sampled
	StaticBackground bool `editable:"Static Background?" edit_type:"bool"`
}

func init() {
	registry.Instance().Register(Sky{})
}

func (m *Sky) Initialize() {
	m.Sampled.Initialize()
	m.Base = m.Sampled.Base
}

func (m *Sky) Deserialize(data map[string]interface{}) {
	m.Initialize()
	m.Base.Deserialize(data)
	if v, ok := data["StaticBackground"]; ok {
		m.StaticBackground = v.(bool)
	}
}
