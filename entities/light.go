package entities

import (
	"github.com/tlyakhov/gofoom/core"
	"github.com/tlyakhov/gofoom/registry"
)

type Light struct {
	core.PhysicalEntity `editable:"^"`
}

func init() {
	registry.Instance().Register(Light{})
}

func (l *Light) Initialize() {
	l.PhysicalEntity.Initialize()
	l.BoundingRadius = 10.0
}

func (l *Light) Serialize() map[string]interface{} {
	result := l.PhysicalEntity.Serialize()
	result["Type"] = "entities.Light"
	return result
}
