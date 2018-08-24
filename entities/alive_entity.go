package entities

import (
	"github.com/tlyakhov/gofoom/core"
	"github.com/tlyakhov/gofoom/registry"
)

type AliveEntity struct {
	core.PhysicalEntity `editable:"^"`
	Health              float64 `editable:"Health"`
	HurtTime            float64
}

func init() {
	registry.Instance().Register(AliveEntity{})
}

func (e *AliveEntity) Initialize() {
	e.PhysicalEntity.Initialize()
	e.Health = 100
}

func (e *AliveEntity) Deserialize(data map[string]interface{}) {
	e.Initialize()
	e.PhysicalEntity.Deserialize(data)

	if v, ok := data["Health"]; ok {
		e.Health = v.(float64)
	}
}

func (e *AliveEntity) Serialize() map[string]interface{} {
	result := e.PhysicalEntity.Serialize()
	result["Type"] = "entities.AliveEntity"
	result["Health"] = e.Health

	return result
}
