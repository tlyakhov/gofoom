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
	LastTurn   int64 // nanoseconds
	LastTarget int64
	NextSector ecs.Entity
}

func (w *Wander) String() string {
	return "Wander"
}

func (w *Wander) Construct(data map[string]any) {
	w.Attached.Construct(data)

	w.Force = 10
	w.LastTurn = ecs.Simulation.SimTimestamp

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
