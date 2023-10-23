package core

import (
	"tlyakhov/gofoom/concepts"
)

type Spawn struct {
	concepts.Attached `editable:"^"`
	Spawn             concepts.Vector3 `editable:"Spawn"`
}

var SpawnComponentIndex int

func init() {
	SpawnComponentIndex = concepts.DbTypes().Register(Spawn{})
}

func SpawnFromDb(entity *concepts.EntityRef) *Spawn {
	return entity.Component(SpawnComponentIndex).(*Spawn)
}

func (w *Spawn) Construct(data map[string]any) {
	w.Attached.Construct(data)
	w.Spawn = concepts.Vector3{}

	if data == nil {
		return
	}

	if v, ok := data["X"]; ok {
		w.Spawn[0] = v.(float64)
	}
	if v, ok := data["Y"]; ok {
		w.Spawn[1] = v.(float64)
	}
}

func (w *Spawn) Serialize() map[string]any {
	result := w.Attached.Serialize()
	result["X"] = w.Spawn[0]
	result["Y"] = w.Spawn[1]
	return result
}
