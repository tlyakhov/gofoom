package mobs

import (
	"tlyakhov/gofoom/core"
	"tlyakhov/gofoom/registry"
)

type AliveMob struct {
	core.PhysicalMob `editable:"^"`
	Health           float64 `editable:"Health"`
	HurtTime         float64
}

func init() {
	registry.Instance().Register(AliveMob{})
}

func (e *AliveMob) Construct(data map[string]interface{}) {
	e.PhysicalMob.Construct(data)
	e.Model = e
	e.Health = 100

	if data == nil {
		return
	}

	if v, ok := data["Health"]; ok {
		e.Health = v.(float64)
	}
}

func (e *AliveMob) Serialize() map[string]interface{} {
	result := e.PhysicalMob.Serialize()
	result["Type"] = "mobs.AliveMob"
	result["Health"] = e.Health

	return result
}
