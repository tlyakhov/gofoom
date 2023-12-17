package materials

import (
	"tlyakhov/gofoom/concepts"
)

type Tiled struct {
	concepts.Attached `editable:"^"`
	IsLiquid          bool `editable:"Is Liquid?" edit_type:"bool"`
}

var TiledComponentIndex int

func init() {
	TiledComponentIndex = concepts.DbTypes().Register(Tiled{}, TiledFromDb)
}

func TiledFromDb(entity *concepts.EntityRef) *Tiled {
	if asserted, ok := entity.Component(TiledComponentIndex).(*Tiled); ok {
		return asserted
	}
	return nil
}

func (m *Tiled) String() string {
	s := "Tiled"
	if m.IsLiquid {
		s += "(Liquid)"
	}
	return s
}

func (m *Tiled) Construct(data map[string]any) {
	m.Attached.Construct(data)

	if data == nil {
		return
	}

	if v, ok := data["IsLiquid"]; ok {
		m.IsLiquid = v.(bool)
	}
}

func (m *Tiled) Serialize() map[string]any {
	result := m.Attached.Serialize()
	result["IsLiquid"] = m.IsLiquid
	return result
}
