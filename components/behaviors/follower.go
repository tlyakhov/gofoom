// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"tlyakhov/gofoom/dynamic"
	"tlyakhov/gofoom/ecs"
)

type Follower struct {
	ecs.Attached `editable:"^"`
	Start        ecs.Entity                `editable:"Starting Action" edit_type:"Action"`
	NoZ          bool                      `editable:"2D only"`
	Lifetime     dynamic.AnimationLifetime `editable:"Lifetime"`
	Speed        float64                   `editable:"Speed"`

	// Internal state
	Action         ecs.Entity
	LastTransition int64
}

var FollowerCID ecs.ComponentID

func init() {
	FollowerCID = ecs.RegisterComponent(&ecs.Column[Follower, *Follower]{Getter: GetFollower}, "FollowerWander")
}

func GetFollower(db *ecs.ECS, e ecs.Entity) *Follower {
	if asserted, ok := db.Component(e, FollowerCID).(*Follower); ok {
		return asserted
	}
	return nil
}

func (f *Follower) String() string {
	return "Follower"
}

func (f *Follower) Construct(data map[string]any) {
	f.Attached.Construct(data)
	f.NoZ = false
	f.Action = f.Start
	f.Speed = 10
	f.Lifetime = dynamic.AnimationLifetimeLoop
	f.LastTransition = 0

	if data == nil {
		return
	}

	if v, ok := data["Action"]; ok {
		f.Action, _ = ecs.ParseEntity(v.(string))
	}
	if v, ok := data["Start"]; ok {
		f.Start, _ = ecs.ParseEntity(v.(string))
	}
	if v, ok := data["NoZ"]; ok {
		f.NoZ = v.(bool)
	}
	if v, ok := data["Lifetime"]; ok {
		if l, err := dynamic.AnimationLifetimeString(v.(string)); err == nil {
			f.Lifetime = l
		}
	}
	if v, ok := data["Speed"]; ok {
		f.Speed = v.(float64)
	}
}

func (f *Follower) Serialize() map[string]any {
	result := f.Attached.Serialize()

	result["Action"] = f.Action.String()
	result["Start"] = f.Start.String()
	result["Speed"] = f.Speed

	if f.NoZ {
		result["NoZ"] = f.NoZ
	}
	if f.Lifetime != dynamic.AnimationLifetimeLoop {
		result["Lifetime"] = f.Lifetime
	}
	return result
}
