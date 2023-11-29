package behaviors

import (
	"fmt"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
)

type Proximity struct {
	concepts.Attached `editable:"^"`
	Range             float64        `editable:"Range"`
	Triggers          []core.Trigger `editable:"Triggers"`
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

func (s *Proximity) String() string {
	return fmt.Sprintf("Proximity: %.2f", s.Range)
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

	if v, ok := data["Triggers"]; ok {
		if triggers, ok := v.([]any); ok {
			s.Triggers = make([]core.Trigger, len(triggers))
			for i, tdata := range triggers {
				s.Triggers[i].Construct(tdata.(map[string]any))
			}
		}
	}
}

func (s *Proximity) Serialize() map[string]any {
	result := s.Attached.Serialize()
	result["Range"] = s.Range
	triggers := make([]map[string]any, len(s.Triggers))
	for i, trigger := range s.Triggers {
		triggers[i] = trigger.Serialize()
	}
	result["Triggers"] = triggers

	return result
}
