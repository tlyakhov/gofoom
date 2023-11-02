package materials

import "tlyakhov/gofoom/concepts"

type Sky struct {
	concepts.Attached `editable:"^"`

	StaticBackground bool `editable:"Static Background?" edit_type:"bool"`
}

var SkyComponentIndex int

func init() {
	SkyComponentIndex = concepts.DbTypes().Register(Sky{})
}

func SkyFromDb(entity *concepts.EntityRef) *Sky {
	if asserted, ok := entity.Component(SkyComponentIndex).(*Sky); ok {
		return asserted
	}
	return nil
}

func (m *Sky) Construct(data map[string]any) {
	m.Attached.Construct(data)

	if data == nil {
		return
	}

	if v, ok := data["StaticBackground"]; ok {
		m.StaticBackground = v.(bool)
	}
}

func (m *Sky) Serialize() map[string]any {
	result := m.Attached.Serialize()
	result["StaticBackground"] = m.StaticBackground
	return result
}
