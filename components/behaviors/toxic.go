package behaviors

import (
	"tlyakhov/gofoom/concepts"
)

type Toxic struct {
	concepts.Attached `editable:"^"`
	Hurt              float64 `editable:"Hurt Amount"`
}

var ToxicComponentIndex int

func init() {
	ToxicComponentIndex = concepts.DbTypes().Register(Toxic{})
}

func ToxicFromDb(entity *concepts.EntityRef) *Toxic {
	if asserted, ok := entity.Component(ToxicComponentIndex).(*Toxic); ok {
		return asserted
	}
	return nil
}

func (s *Toxic) Construct(data map[string]any) {
	s.Attached.Construct(data)

	if data == nil {
		return
	}

	if v, ok := data["Hurt"]; ok {
		s.Hurt = v.(float64)
	}
}

func (s *Toxic) Serialize() map[string]any {
	result := s.Attached.Serialize()
	result["Hurt"] = s.Hurt
	return result
}
