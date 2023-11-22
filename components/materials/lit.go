package materials

import (
	"tlyakhov/gofoom/concepts"
)

type Lit struct {
	concepts.Attached `editable:"^"`

	Ambient concepts.Vector3 `editable:"Ambient Color" edit_type:"color"`
	Diffuse concepts.Vector3 `editable:"Diffuse Color" edit_type:"color"`
}

var LitComponentIndex int

func init() {
	LitComponentIndex = concepts.DbTypes().Register(Lit{})
}

func LitFromDb(entity *concepts.EntityRef) *Lit {
	if asserted, ok := entity.Component(LitComponentIndex).(*Lit); ok {
		return asserted
	}
	return nil
}

func (m *Lit) String() string {
	return "Lit: " + m.Diffuse.String()
}

func (m *Lit) Construct(data map[string]any) {
	m.Attached.Construct(data)

	m.Ambient = concepts.Vector3{0.1, 0.1, 0.1}
	m.Diffuse = concepts.Vector3{1, 1, 1}

	if data == nil {
		return
	}

	if v, ok := data["Ambient"]; ok {
		m.Ambient.Deserialize(v.(map[string]any))
	}
	if v, ok := data["Diffuse"]; ok {
		m.Diffuse.Deserialize(v.(map[string]any))
	}
}

func (m *Lit) Serialize() map[string]any {
	result := m.Attached.Serialize()
	result["Ambient"] = m.Ambient.Serialize()
	result["Diffuse"] = m.Diffuse.Serialize()
	return result
}
