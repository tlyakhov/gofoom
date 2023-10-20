package entities

import (
	"tlyakhov/gofoom/core"
	"tlyakhov/gofoom/registry"
)

type AliveEntity struct {
	core.PhysicalEntity `editable:"^"`
	Health              float64 `editable:"Health"`
	HurtTime            float64
}

func init() {
	registry.Instance().Register(AliveEntity{})
}

func (e *AliveEntity) Construct(data map[string]interface{}) {
	e.PhysicalEntity.Construct(data)
	e.Model = e
	e.Health = 100

	if data == nil {
		return
	}

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
