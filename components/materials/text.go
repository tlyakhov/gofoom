package materials

import "tlyakhov/gofoom/concepts"

type Text struct {
	concepts.Attached `editable:"^"`

	Label string `editable:"Label"`
}

var TextComponentIndex int

func init() {
	TextComponentIndex = concepts.DbTypes().Register(Text{}, TextFromDb)
}

func TextFromDb(entity *concepts.EntityRef) *Text {
	if asserted, ok := entity.Component(TextComponentIndex).(*Text); ok {
		return asserted
	}
	return nil
}

func (m *Text) Construct(data map[string]any) {
	m.Attached.Construct(data)

	if data == nil {
		return
	}

	if v, ok := data["Label"]; ok {
		m.Label = v.(string)
	}
}

func (m *Text) Serialize() map[string]any {
	result := m.Attached.Serialize()
	result["Label"] = m.Label
	return result
}
