package behaviors

import (
	"tlyakhov/gofoom/concepts"
)

type Alive struct {
	concepts.Attached `editable:"^"`
	Health            float64 `editable:"Health"`
	HurtTime          float64
}

var AliveComponentIndex int

func init() {
	AliveComponentIndex = concepts.DbTypes().Register(Alive{})
}

func AliveFromDb(entity *concepts.EntityRef) *Alive {
	return entity.Component(AliveComponentIndex).(*Alive)
}

func (e *Alive) Construct(data map[string]any) {
	e.Attached.Construct(data)

	e.Health = 100

	if data == nil {
		return
	}

	if v, ok := data["Health"]; ok {
		e.Health = v.(float64)
	}
}

func (e *Alive) Serialize() map[string]any {
	result := e.Attached.Serialize()
	result["Health"] = e.Health

	return result
}
