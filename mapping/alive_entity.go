package mapping

import "github.com/tlyakhov/gofoom/registry"

type AliveEntity struct {
	Entity
	Health   float64
	HurtTime float64
}

func init() {
	registry.Instance().Register(AliveEntity{})
}

func (e *AliveEntity) Initialize() {
	e.Entity.Initialize()
	e.Health = 100
}

func (e *AliveEntity) Deserialize(data map[string]interface{}) {
	e.Initialize()
	e.Entity.Deserialize(data)

	if v, ok := data["Health"]; ok {
		e.Health = v.(float64)
	}
}
