package behaviors

import (
	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/core"
	"github.com/tlyakhov/gofoom/registry"
)

type Light struct {
	core.AnimatedBehavior `editable:"^"`

	Diffuse     concepts.Vector3 `editable:"Diffuse"`
	Strength    float64          `editable:"Strength"`
	Attenuation float64          `editable:"Attenuation"`
	LastPos     concepts.Vector3
}

func init() {
	registry.Instance().Register(Light{})
}

func (l *Light) Initialize() {
	l.AnimatedBehavior.Initialize()

	l.Diffuse = concepts.Vector3{1, 1, 1}
	l.Strength = 15
	l.Attenuation = 1.2
}

func (l *Light) Frame(lastFrameTime float64) {
	l.AnimatedBehavior.Frame(lastFrameTime)

	if !l.Active {
		return
	}

	pos := l.Entity.Physical().Pos
	if pos != l.LastPos {
		l.Entity.Physical().Map.ClearLightmaps()
	}
	l.LastPos = pos
}

func (l *Light) Deserialize(data map[string]interface{}) {
	l.Initialize()
	l.AnimatedBehavior.Deserialize(data)

	if v, ok := data["Diffuse"]; ok {
		l.Diffuse.Deserialize(v.(map[string]interface{}))
	}
	if v, ok := data["Strength"]; ok {
		l.Strength = v.(float64)
	}
	if v, ok := data["Attenuation"]; ok {
		l.Attenuation = v.(float64)
	}
}

func (l *Light) Serialize() map[string]interface{} {
	result := l.AnimatedBehavior.Serialize()
	result["Type"] = "behaviors.Light"
	result["Diffuse"] = l.Diffuse.Serialize()
	result["Strength"] = l.Strength
	result["Attenuation"] = l.Attenuation

	return result
}
