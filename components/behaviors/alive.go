package behaviors

import (
	"fmt"
	"tlyakhov/gofoom/concepts"
)

type Alive struct {
	concepts.Attached `editable:"^"`
	Health            float64 `editable:"Health"`
	HurtTime          float64
}

var AliveComponentIndex int

func init() {
	AliveComponentIndex = concepts.DbTypes().Register(Alive{}, AliveFromDb)
}

func AliveFromDb(entity *concepts.EntityRef) *Alive {
	if asserted, ok := entity.Component(AliveComponentIndex).(*Alive); ok {
		return asserted
	}
	return nil
}

func (a *Alive) String() string {
	return fmt.Sprintf("Alive: %.2f", a.Health)
}

func (a *Alive) Construct(data map[string]any) {
	a.Attached.Construct(data)

	a.Health = 100

	if data == nil {
		return
	}

	if v, ok := data["Health"]; ok {
		a.Health = v.(float64)
	}
}

func (a *Alive) Serialize() map[string]any {
	result := a.Attached.Serialize()
	result["Health"] = a.Health

	return result
}
