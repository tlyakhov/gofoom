package materials

import (
	"tlyakhov/gofoom/registry"
)

type Sky struct {
	Sampled `editable:"^"`

	StaticBackground bool `editable:"Static Background?" edit_type:"bool"`
}

func init() {
	registry.Instance().Register(Sky{})
}

func (m *Sky) Construct(data map[string]interface{}) {
	m.Sampled.Construct(data)
	m.Model = m

	if data == nil {
		return
	}

	if v, ok := data["StaticBackground"]; ok {
		m.StaticBackground = v.(bool)
	}
}

func (m *Sky) Serialize() map[string]interface{} {
	result := m.Sampled.Serialize()
	result["Type"] = "materials.Sky"
	result["StaticBackground"] = m.StaticBackground
	return result
}
