package materials

import (
	"tlyakhov/gofoom/concepts"
)

type Sprite struct {
	DB *concepts.EntityComponentDB

	Image *concepts.EntityRef `editable:"Image" edit_type:"Material"`
	Frame int                 `editable:"Frame"`
	Angle int                 `editable:"Angle"`
}

type SpriteSheet struct {
	concepts.Attached `editable:"^"`

	Sprites []*Sprite `editable:"Sprites"`
}

var SpriteSheetComponentIndex int

func init() {
	SpriteSheetComponentIndex = concepts.DbTypes().Register(SpriteSheet{}, SpriteSheetFromDb)
}

func SpriteSheetFromDb(entity *concepts.EntityRef) *SpriteSheet {
	if asserted, ok := entity.Component(SpriteSheetComponentIndex).(*SpriteSheet); ok {
		return asserted
	}
	return nil
}

func (s *SpriteSheet) Construct(data map[string]any) {
	s.Attached.Construct(data)

	if data == nil {
		return
	}

	if v, ok := data["Sprites"]; ok {
		s.Sprites = concepts.ConstructSlice[*Sprite](s.DB, v)
	}
}

func (s *SpriteSheet) Serialize() map[string]any {
	result := s.Attached.Serialize()
	if s.Sprites != nil {
		result["Sprites"] = concepts.SerializeSlice(s.Sprites)
	}

	return result
}

func (s *Sprite) Construct(data map[string]any) {
	s.Frame = 0
	s.Angle = 0

	if data == nil {
		return
	}
	if v, ok := data["Image"]; ok {
		s.Image = s.DB.DeserializeEntityRef(v)
	}
	if v, ok := data["Frame"]; ok {
		s.Frame = int(v.(float64))
	}
	if v, ok := data["Angle"]; ok {
		s.Angle = int(v.(float64))
	}
}

func (s *Sprite) Serialize() map[string]any {
	result := make(map[string]any)
	if !s.Image.Nil() {
		result["Image"] = s.Image.Serialize()
	}
	result["Frame"] = s.Frame
	result["Angle"] = s.Angle
	return result
}

func (s *Sprite) SetDB(db *concepts.EntityComponentDB) {
	s.DB = db
}
