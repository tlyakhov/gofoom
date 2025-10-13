// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"tlyakhov/gofoom/dynamic"
	"tlyakhov/gofoom/ecs"

	"github.com/spf13/cast"
)

type Actor struct {
	ecs.Attached `editable:"^"`
	Start        ecs.Entity                `editable:"Starting Action" edit_type:"Action"`
	NoZ          bool                      `editable:"2D only"`
	Lifetime     dynamic.AnimationLifetime `editable:"Lifetime"`

	// Units per second
	Speed float64 `editable:"Speed"`
	// Degrees per second
	AngularSpeed float64 `editable:"Angular Speed"`

	FaceNextWaypoint bool `editable:"Face Waypoint?"`
}

func (a *Actor) String() string {
	return "Actor"
}

func (a *Actor) Construct(data map[string]any) {
	a.Attached.Construct(data)
	a.NoZ = false
	a.Speed = 10
	a.AngularSpeed = 15
	a.Lifetime = dynamic.AnimationLifetimeLoop
	a.FaceNextWaypoint = true

	if data == nil {
		return
	}

	if v, ok := data["Start"]; ok {
		a.Start, _ = ecs.ParseEntity(v.(string))
	}
	if v, ok := data["NoZ"]; ok {
		a.NoZ = cast.ToBool(v)
	}
	if v, ok := data["Lifetime"]; ok {
		if l, err := dynamic.AnimationLifetimeString(v.(string)); err == nil {
			a.Lifetime = l
		}
	}
	if v, ok := data["Speed"]; ok {
		a.Speed = cast.ToFloat64(v)
	}
	if v, ok := data["AngularSpeed"]; ok {
		a.AngularSpeed = cast.ToFloat64(v)
	}
	if v, ok := data["FaceNextWaypoint"]; ok {
		a.FaceNextWaypoint = v.(bool)
	}

}

func (a *Actor) Serialize() map[string]any {
	result := a.Attached.Serialize()

	result["Start"] = a.Start.Serialize()
	result["Speed"] = a.Speed
	result["AngularSpeed"] = a.AngularSpeed

	if a.NoZ {
		result["NoZ"] = a.NoZ
	}
	if a.Lifetime != dynamic.AnimationLifetimeLoop {
		result["Lifetime"] = a.Lifetime
	}
	if !a.FaceNextWaypoint {
		result["FaceNextWaypoint"] = a.FaceNextWaypoint
	}
	return result
}
