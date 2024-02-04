package core

import (
	"tlyakhov/gofoom/concepts"
)

type Light struct {
	concepts.Attached `editable:"^"`

	Diffuse     concepts.Vector3 `editable:"Diffuse"`
	Strength    float64          `editable:"Strength"`
	Attenuation float64          `editable:"Attenuation"`
}

var LightComponentIndex int

func init() {
	LightComponentIndex = concepts.DbTypes().Register(Light{}, LightFromDb)
}

func LightFromDb(entity *concepts.EntityRef) *Light {
	if asserted, ok := entity.Component(LightComponentIndex).(*Light); ok {
		return asserted
	}
	return nil
}

func (l *Light) String() string {
	return "Light: " + l.Diffuse.StringHuman()
}

func (l *Light) Construct(data map[string]any) {
	l.Attached.Construct(data)
	l.Diffuse = concepts.Vector3{1, 1, 1}
	l.Strength = 2
	l.Attenuation = 0.4

	if data == nil {
		return
	}

	if v, ok := data["Diffuse"]; ok {
		l.Diffuse.Deserialize(v.(map[string]any))
	}
	if v, ok := data["Strength"]; ok {
		l.Strength = v.(float64)
	}
	if v, ok := data["Attenuation"]; ok {
		l.Attenuation = v.(float64)
	}
}

func (l *Light) Serialize() map[string]any {
	result := l.Attached.Serialize()
	result["Diffuse"] = l.Diffuse.Serialize()
	result["Strength"] = l.Strength
	result["Attenuation"] = l.Attenuation

	return result
}
