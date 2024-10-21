// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"tlyakhov/gofoom/dynamic"
	"tlyakhov/gofoom/ecs"
)

type Actor struct {
	ecs.Attached `editable:"^"`
	Start        ecs.Entity                `editable:"Starting Action" edit_type:"Action"`
	NoZ          bool                      `editable:"2D only"`
	Lifetime     dynamic.AnimationLifetime `editable:"Lifetime"`
	Speed        float64                   `editable:"Speed"`
}

var ActorCID ecs.ComponentID

func init() {
	ActorCID = ecs.RegisterComponent(&ecs.Column[Actor, *Actor]{Getter: GetActor})
}

func GetActor(db *ecs.ECS, e ecs.Entity) *Actor {
	if asserted, ok := db.Component(e, ActorCID).(*Actor); ok {
		return asserted
	}
	return nil
}

func (a *Actor) String() string {
	return "Actor"
}

func (a *Actor) Construct(data map[string]any) {
	a.Attached.Construct(data)
	a.NoZ = false
	a.Speed = 10
	a.Lifetime = dynamic.AnimationLifetimeLoop

	if data == nil {
		return
	}

	if v, ok := data["Start"]; ok {
		a.Start, _ = ecs.ParseEntity(v.(string))
	}
	if v, ok := data["NoZ"]; ok {
		a.NoZ = v.(bool)
	}
	if v, ok := data["Lifetime"]; ok {
		if l, err := dynamic.AnimationLifetimeString(v.(string)); err == nil {
			a.Lifetime = l
		}
	}
	if v, ok := data["Speed"]; ok {
		a.Speed = v.(float64)
	}
}

func (a *Actor) Serialize() map[string]any {
	result := a.Attached.Serialize()

	result["Start"] = a.Start.String()
	result["Speed"] = a.Speed

	if a.NoZ {
		result["NoZ"] = a.NoZ
	}
	if a.Lifetime != dynamic.AnimationLifetimeLoop {
		result["Lifetime"] = a.Lifetime
	}
	return result
}
