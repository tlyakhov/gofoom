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

	Sprites []Sprite `editable:"Sprites"`
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
		s.Sprites = ConstructSprites(s.DB, v)
	}
}

func (s *SpriteSheet) Serialize() map[string]any {
	result := s.Attached.Serialize()
	if s.Sprites != nil {
		result["Sprites"] = SerializeSprites(s.Sprites)
	}
	return result
}

func ConstructSprites(db *concepts.EntityComponentDB, data any) []Sprite {
	var result []Sprite

	if Sprites, ok := data.([]any); ok {
		result = make([]Sprite, len(Sprites))
		for i, tdata := range Sprites {
			result[i].Construct(db, tdata.(map[string]any))
		}
	}
	return result
}

func SerializeSprites(Sprites []Sprite) []map[string]any {
	result := make([]map[string]any, len(Sprites))
	for i, Sprite := range Sprites {
		result[i] = Sprite.Serialize()
	}
	return result
}

func (s *Sprite) Construct(db *concepts.EntityComponentDB, data map[string]any) {
	s.DB = db
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
