package entities

import (
	"tlyakhov/gofoom/behaviors"
	"tlyakhov/gofoom/core"
	"tlyakhov/gofoom/registry"
)

type Light struct {
	core.PhysicalEntity `editable:"^"`
}

func init() {
	registry.Instance().Register(Light{})
}

func (l *Light) Serialize() map[string]interface{} {
	result := l.PhysicalEntity.Serialize()
	result["Type"] = "entities.Light"
	return result
}

func (l *Light) Construct(data map[string]interface{}) {
	l.PhysicalEntity.Construct(data)
	l.Model = l
	l.BoundingRadius = 10.0

	lb := &behaviors.Light{}
	lb.Construct(nil)
	lb.ID = "Light"
	l.Behaviors[lb.GetBase().ID] = lb
	lb.SetParent(l)
}
