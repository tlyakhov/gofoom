package materials

import (
	"tlyakhov/gofoom/concepts"
)

type Tiled struct {
	concepts.Attached `editable:"^"`
	IsLiquid          bool    `editable:"Is Liquid?" edit_type:"bool"`
	Scale             float64 `editable:"Scale" edit_type:"float"`
}

var TiledComponentIndex int

func init() {
	TiledComponentIndex = concepts.DbTypes().Register(Tiled{})
}

func TiledFromDb(entity *concepts.EntityRef) *Tiled {
	return entity.Component(TiledComponentIndex).(*Tiled)
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
