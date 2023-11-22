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

func (t *Toxic) String() string {
	return "Toxic"
}

func (t *Toxic) Construct(data map[string]any) {
	t.Attached.Construct(data)

	if data == nil {
		return
	}

	if v, ok := data["Hurt"]; ok {
		t.Hurt = v.(float64)
	}
}

func (t *Toxic) Serialize() map[string]any {
	result := t.Attached.Serialize()
	result["Hurt"] = t.Hurt
	return result
}
