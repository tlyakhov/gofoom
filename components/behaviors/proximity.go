package behaviors

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
)

type Proximity struct {
	concepts.Attached `editable:"^"`
	Range             float64         `editable:"Range"`
	Condition         core.Expression `editable:"Trigger Condition"`
	Trigger           core.Expression `editable:"Trigger"`
}

var ProximityComponentIndex int

func init() {
	ProximityComponentIndex = concepts.DbTypes().Register(Proximity{}, ProximityFromDb)
}

func ProximityFromDb(entity *concepts.EntityRef) *Proximity {
	if asserted, ok := entity.Component(ProximityComponentIndex).(*Proximity); ok {
		return asserted
	}
	return nil
}

func (s *Proximity) Construct(data map[string]any) {
	s.Attached.Construct(data)

	s.Range = 100

	if data == nil {
		return
	}

	if v, ok := data["Range"]; ok {
		s.Range = v.(float64)
	}

	if v, ok := data["Condition"]; ok {
		s.Condition.Construct(v.(string))
	}
	if v, ok := data["Trigger"]; ok {
		s.Trigger.Construct(v.(string))
	}
}

func (s *Proximity) Serialize() map[string]any {
	result := s.Attached.Serialize()
	result["Range"] = s.Range
	result["Condition"] = s.Condition.Serialize()
	result["Trigger"] = s.Trigger.Serialize()

	return result
}
