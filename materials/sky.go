package materials

import (
	"strconv"

	"github.com/tlyakhov/gofoom/registry"
)

type Sky struct {
	Sampled
	StaticBackground bool `editable:"Static Background?" edit_type:"bool"`
}

func init() {
	registry.Instance().Register(Sky{})
}

func (m *Sky) Initialize() {
	m.Sampled.Initialize()
}

func (m *Sky) Deserialize(data map[string]interface{}) {
	m.Initialize()
	m.Sampled.Deserialize(data)
	if v, ok := data["StaticBackground"]; ok {
		m.StaticBackground = v.(bool)
	}
}

func (m *Sky) Serialize() map[string]interface{} {
	result := m.Sampled.Serialize()
	result["Type"] = "materials.Sky"
	result["StaticBackground"] = strconv.FormatBool(m.StaticBackground)
	return result
}
