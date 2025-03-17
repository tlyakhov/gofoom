// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"tlyakhov/gofoom/ecs"

	"github.com/spf13/cast"
)

type Wander struct {
	ecs.Attached `editable:"^"`

	Force     float64 `editable:"Force"`
	AsImpulse bool    `editable:"Apply as impulse"`

	// Internal state
	LastTurn   int64
	LastTarget int64
	NextSector ecs.Entity
}

var WanderCID ecs.ComponentID

func init() {
	WanderCID = ecs.RegisterComponent(&ecs.Column[Wander, *Wander]{Getter: GetWander})
}

func GetWander(u *ecs.Universe, e ecs.Entity) *Wander {
	if asserted, ok := u.Component(e, WanderCID).(*Wander); ok {
		return asserted
	}
	return nil
}

func (w *Wander) String() string {
	return "Wander"
}

func (w *Wander) Construct(data map[string]any) {
	w.Attached.Construct(data)

	w.Force = 10
	w.LastTurn = w.Universe.Timestamp

	if data == nil {
		return
	}

	if v, ok := data["Force"]; ok {
		w.Force = cast.ToFloat64(v)
	}
	if v, ok := data["AsImpulse"]; ok {
		w.AsImpulse = cast.ToBool(v)
	}
}

func (w *Wander) Serialize() map[string]any {
	result := w.Attached.Serialize()

	result["Force"] = w.Force
	if w.AsImpulse {
		result["AsImpulse"] = true
	}
	return result
}
