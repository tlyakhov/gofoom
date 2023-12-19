package behaviors

import (
	"tlyakhov/gofoom/concepts"
)

type Wander struct {
	concepts.Attached `editable:"^"`

	Force      float64 `editable:"Force"`
	Dir        concepts.Vector3
	LastChange int64
}

var WanderComponentIndex int

func init() {
	WanderComponentIndex = concepts.DbTypes().Register(Wander{}, WanderFromDb)
}

func WanderFromDb(entity *concepts.EntityRef) *Wander {
	if asserted, ok := entity.Component(WanderComponentIndex).(*Wander); ok {
		return asserted
	}
	return nil
}

func (w *Wander) String() string {
	return "Wander"
}

func (w *Wander) Construct(data map[string]any) {
	w.Attached.Construct(data)

	w.Force = 10
	w.LastChange = w.DB.Timestamp

	if data == nil {
		return
	}

	if v, ok := data["Force"]; ok {
		w.Force = v.(float64)
	}

}

func (w *Wander) Serialize() map[string]any {
	result := w.Attached.Serialize()

	result["Force"] = w.Force
	return result
}
