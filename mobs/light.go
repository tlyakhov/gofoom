package mobs

import (
	"tlyakhov/gofoom/behaviors"
	"tlyakhov/gofoom/core"
	"tlyakhov/gofoom/registry"
)

type Light struct {
	core.PhysicalMob `editable:"^"`
}

func init() {
	registry.Instance().Register(Light{})
}

func (l *Light) Serialize() map[string]interface{} {
	result := l.PhysicalMob.Serialize()
	result["Type"] = "mobs.Light"
	return result
}

func (l *Light) Construct(data map[string]interface{}) {
	l.PhysicalMob.Construct(data)
	l.Model = l
	l.BoundingRadius = 10.0
	if data == nil {
		lb := &behaviors.Light{}
		lb.Construct(data)
		lb.Name = "Light"
		l.Behaviors[lb.GetBase().Name] = lb
		lb.SetParent(l)
	}
}
