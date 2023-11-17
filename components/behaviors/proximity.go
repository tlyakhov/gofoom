package behaviors

import (
	"tlyakhov/gofoom/concepts"
)

type Proximity struct {
	concepts.Attached `editable:"^"`
	Range             float64      `editable:"Range"`
	TriggerSources    map[int]bool `editable:"Triggering Components"`
	TargetTriggers    map[int]bool `editable:"Components to Trigger"`
}

var ProximityComponentIndex int

func init() {
	ProximityComponentIndex = concepts.DbTypes().Register(Proximity{})
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

	s.DeserializeComponentList(&s.TriggerSources, "TriggerSources", data)
	s.DeserializeComponentList(&s.TargetTriggers, "TargetTriggers", data)
}

func (s *Proximity) Serialize() map[string]any {
	result := s.Attached.Serialize()
	result["Range"] = s.Range
	s.SerializeComponentList(s.TriggerSources, "TriggerSources", result)
	s.SerializeComponentList(s.TargetTriggers, "TargetTriggers", result)
	return result
}
