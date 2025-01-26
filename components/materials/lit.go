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
	LitCID = ecs.RegisterComponent(&ecs.Column[Lit, *Lit]{Getter: GetLit})
}

func GetLit(db *ecs.ECS, e ecs.Entity) *Lit {
	if asserted, ok := db.Component(e, LitCID).(*Lit); ok {
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
		light.To3D().AddSelf(&m.Ambient)
		light.Mul4Self(&m.Diffuse)
		return result.Mul4Self(light)
	} else {
		// result = Surface * (Diffuse + Ambient)
		a := m.Diffuse
		a.To3D().AddSelf(&m.Ambient)
		return result.Mul4Self(&a)
	}
}
