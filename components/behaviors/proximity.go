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

func (p *Proximity) String() string {
	return fmt.Sprintf("Proximity: %.2f", p.Range)
}

func (p *Proximity) Construct(data map[string]any) {
	p.Attached.Construct(data)

	p.Range = 100

	if data == nil {
		return
	}

	if v, ok := data["Range"]; ok {
		p.Range = v.(float64)
	}

	if v, ok := data["Triggers"]; ok {
		p.Triggers = core.ConstructTriggers(p.DB, v)
	}

}

func (p *Proximity) Serialize() map[string]any {
	result := p.Attached.Serialize()
	result["Range"] = p.Range

	if len(p.Triggers) > 0 {
		result["Triggers"] = core.SerializeTriggers(p.Triggers)
	}

	return result
}
