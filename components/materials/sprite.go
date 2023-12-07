package materials

import (
	"tlyakhov/gofoom/concepts"
)

type Sprite struct {
	concepts.Attached `editable:"^"`
	Texture           *concepts.EntityRef `editable:"Texture" edit_type:"Material"`
	Frame             int
	Angle             int
}

var SpriteComponentIndex int

func init() {
	SpriteComponentIndex = concepts.DbTypes().Register(Sprite{}, SpriteFromDb)
}

func SpriteFromDb(entity *concepts.EntityRef) *Sprite {
	if asserted, ok := entity.Component(SpriteComponentIndex).(*Sprite); ok {
		return asserted
	}
	return nil
}

func (s *Sprite) Construct(data map[string]any) {
	s.Attached.Construct(data)

	if data == nil {
		return
	}
	if v, ok := data["Texture"]; ok {
		s.Texture = s.DB.DeserializeEntityRef(v)
	}
	if v, ok := data["Frame"]; ok {
		s.Frame = v.(int)
	}
	if v, ok := data["Angle"]; ok {
		s.Angle = v.(int)
	}
}

func (s *Sprite) Serialize() map[string]any {
	result := s.Attached.Serialize()
	if !s.Texture.Nil() {
		result["Texture"] = s.Texture.Serialize()
	}
	result["Frame"] = s.Frame
	result["Angle"] = s.Angle
	return result
}
