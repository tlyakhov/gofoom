// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"tlyakhov/gofoom/ecs"
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

var WanderComponentIndex int

func init() {
	WanderComponentIndex = ecs.Types().Register(Wander{}, WanderFromDb)
}

func WanderFromDb(db *ecs.ECS, e ecs.Entity) *Wander {
	if asserted, ok := db.Component(e, WanderComponentIndex).(*Wander); ok {
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
	w.LastTurn = w.DB.Timestamp

	if data == nil {
		return
	}

	if v, ok := data["Force"]; ok {
		w.Force = v.(float64)
	}
	if v, ok := data["AsImpulse"]; ok {
		w.AsImpulse = v.(bool)
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
