package materials

import (
	"fmt"
	"image/color"

	"tlyakhov/gofoom/concepts"
)

type Solid struct {
	concepts.Attached `editable:"^"`
	Diffuse           color.NRGBA
}

var SolidComponentIndex int

func init() {
	SolidComponentIndex = concepts.DbTypes().Register(Solid{})
}

func SolidFromDb(entity *concepts.EntityRef) *Solid {
	return entity.Component(SolidComponentIndex).(*Solid)
}

func (s *Solid) Construct(data map[string]any) {
	s.Attached.Construct(data)

	if data == nil {
		return
	}

	if v, ok := data["Diffuse"]; ok {
		var err error
		s.Diffuse, err = concepts.ParseHexColor(v.(string))
		if err != nil {
			fmt.Printf("Error deserializing Solid texture: %v\n", err)
		}
	}
}

func (s *Solid) Serialize() map[string]any {
	result := s.Attached.Serialize()
	if s.Diffuse.A == 255 {
		result["Diffuse"] = fmt.Sprintf("#%02X%02X%02X", s.Diffuse.R, s.Diffuse.G, s.Diffuse.B)
	} else {
		result["Diffuse"] = fmt.Sprintf("#%02X%02X%02X%02X", s.Diffuse.R, s.Diffuse.G, s.Diffuse.B, s.Diffuse.A)
	}
	return result
}
