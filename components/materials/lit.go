// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package materials

import (
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/ecs"
)

type Lit struct {
	ecs.Attached `editable:"^"`

	Ambient concepts.Vector3 `editable:"Ambient Color" edit_type:"color"`
	Diffuse concepts.Vector4 `editable:"Diffuse Color" edit_type:"color"`
}

var LitCID ecs.ComponentID

func init() {
	LitCID = ecs.RegisterComponent(&ecs.Arena[Lit, *Lit]{})
}

func (x *Lit) ComponentID() ecs.ComponentID {
	return LitCID
}
func GetLit(e ecs.Entity) *Lit {
	if asserted, ok := ecs.Component(e, LitCID).(*Lit); ok {
		return asserted
	}
	return nil
}

func (m *Lit) MultiAttachable() bool { return true }

func (m *Lit) String() string {
	return "Lit: " + m.Diffuse.String()
}

func (m *Lit) Construct(data map[string]any) {
	m.Attached.Construct(data)

	m.Ambient = concepts.Vector3{0, 0, 0}
	m.Diffuse = concepts.Vector4{1, 1, 1, 1}

	if data == nil {
		return
	}

	if v, ok := data["Ambient"]; ok {
		m.Ambient.Deserialize(v.(string))
	}
	if v, ok := data["Diffuse"]; ok {
		m.Diffuse.Deserialize(v.(string))
	}
}

func (m *Lit) Serialize() map[string]any {
	result := m.Attached.Serialize()
	result["Ambient"] = m.Ambient.Serialize()
	result["Diffuse"] = m.Diffuse.Serialize(true)
	return result
}

func (m *Lit) Apply(result, light *concepts.Vector4) *concepts.Vector4 {
	if light != nil {
		// result = Surface * Diffuse * (Ambient + Lightmap)
		light[2] = m.Diffuse[2] * (m.Ambient[2] + light[2])
		light[1] = m.Diffuse[1] * (m.Ambient[1] + light[1])
		light[0] = m.Diffuse[0] * (m.Ambient[0] + light[0])
		result[3] *= m.Diffuse[3]
		result[2] *= light[2]
		result[1] *= light[1]
		result[0] *= light[0]
		return result
	} else {
		// result = Surface * (Diffuse + Ambient)
		a := m.Diffuse
		a.To3D().AddSelf(&m.Ambient)
		return result.Mul4Self(&a)
	}
}
