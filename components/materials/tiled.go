package materials

import (
	"fmt"
	"tlyakhov/gofoom/concepts"
)

type Tiled struct {
	concepts.Attached `editable:"^"`
	IsLiquid          bool    `editable:"Is Liquid?" edit_type:"bool"`
	Scale             float64 `editable:"Scale" edit_type:"float"`
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
	s += fmt.Sprintf(": %.2f", m.Scale)
	return s
}

func (m *Tiled) Construct(data map[string]any) {
	m.Attached.Construct(data)

	m.Scale = 1.0

	if data == nil {
		return
	}

	if v, ok := data["IsLiquid"]; ok {
		m.IsLiquid = v.(bool)
	}
	if v, ok := data["Scale"]; ok {
		m.Scale = v.(float64)
	}

}

func (m *Tiled) Serialize() map[string]any {
	result := m.Attached.Serialize()
	result["IsLiquid"] = m.IsLiquid
	result["Scale"] = m.Scale
	return result
}
