// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package core

import (
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/ecs"

	"github.com/spf13/cast"
)

type Light struct {
	ecs.Attached `editable:"^"`

	Diffuse     concepts.Vector3 `editable:"Diffuse"`
	Strength    float64          `editable:"Strength"`
	Attenuation float64          `editable:"Attenuation"`
}

var LightCID ecs.ComponentID

func init() {
	LightCID = ecs.RegisterComponent(&ecs.Arena[Light, *Light]{})
}

func (x *Light) ComponentID() ecs.ComponentID {
	return LightCID
}
func GetLight(e ecs.Entity) *Light {
	if asserted, ok := ecs.Component(e, LightCID).(*Light); ok {
		return asserted
	}
	return nil
}

func (l *Light) OnDetach(e ecs.Entity) {
	defer l.Attached.OnDetach(e)
	if !l.IsAttached() {
		return
	}

	if b := GetBody(e); b != nil && b.QuadNode != nil {
		b.QuadNode.Remove(b)
		b.QuadNode = nil

	}
}

func (l *Light) OnDelete() {
	defer l.Attached.OnDelete()
	if l.IsAttached() {
		for _, e := range l.Entities {
			if e == 0 {
				continue
			}
			if b := GetBody(e); b != nil && b.QuadNode != nil {
				b.QuadNode.Remove(b)
				b.QuadNode = nil
			}
		}
	}
}

func (l *Light) OnAttach() {
	l.Attached.OnAttach()

	if tree := ecs.Singleton(QuadtreeCID).(*Quadtree); tree != nil {
		for _, e := range l.Entities {
			if e == 0 {
				continue
			}
			if b := GetBody(e); b != nil && b.QuadNode != nil {
				b.QuadNode.Lights = append(b.QuadNode.Lights, b)
			}
		}
	}
}

func (l *Light) MultiAttachable() bool { return true }

func (l *Light) String() string {
	return "Light: " + l.Diffuse.StringHuman(2)
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
		l.Diffuse.Deserialize(v.(string))
	}
	if v, ok := data["Strength"]; ok {
		l.Strength = cast.ToFloat64(v)
	}
	if v, ok := data["Attenuation"]; ok {
		l.Attenuation = cast.ToFloat64(v)
	}
}

func (l *Light) Serialize() map[string]any {
	result := l.Attached.Serialize()
	result["Diffuse"] = l.Diffuse.Serialize()
	result["Strength"] = l.Strength
	result["Attenuation"] = l.Attenuation

	return result
}
