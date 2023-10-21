package behaviors

import (
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/registry"
)

type Light struct {
	AnimatedBehavior `editable:"^"`

	Diffuse     concepts.Vector3 `editable:"Diffuse"`
	Strength    float64          `editable:"Strength"`
	Attenuation float64          `editable:"Attenuation"`
	LastPos     concepts.Vector3
}

func init() {
	registry.Instance().Register(Light{})
}

func (l *Light) Frame() {
	l.AnimatedBehavior.Frame()

	if !l.Active {
		return
	}
}

func (l *Light) Construct(data map[string]interface{}) {
	l.AnimatedBehavior.Construct(data)
	l.Model = l
	l.Diffuse = concepts.Vector3{1, 1, 1}
	l.Strength = 2
	l.Attenuation = 0.4

	if data == nil {
		return
	}

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
