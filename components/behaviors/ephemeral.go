// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"tlyakhov/gofoom/ecs"

	"github.com/spf13/cast"
)

type Ephemeral struct {
	ecs.Attached `editable:"^"`

	Lifetime             float64 `editable:"Lifetime"`  // ms
	FadeTime             float64 `editable:"Fade Time"` //ms
	DeleteEntityOnExpiry bool    `editable:"Delete Entity on Expiry"`

	CreationTime int64 // ns
}

func (e *Ephemeral) Shareable() bool { return true }

func (e *Ephemeral) String() string {
	return "Ephemeral"
}

func (e *Ephemeral) Construct(data map[string]any) {
	e.Attached.Construct(data)

	e.Lifetime = 60000 // 1min by default
	e.FadeTime = 1000  // 1s by default
	e.DeleteEntityOnExpiry = false
	e.CreationTime = ecs.Simulation.SimTimestamp

	if data == nil {
		return
	}

	if v, ok := data["Lifetime"]; ok {
		e.Lifetime = cast.ToFloat64(v)
	}
	if v, ok := data["Lifetime"]; ok {
		e.Lifetime = cast.ToFloat64(v)
	}
	if v, ok := data["DeleteEntityOnExpiry"]; ok {
		e.DeleteEntityOnExpiry = cast.ToBool(v)
	}
}

func (e *Ephemeral) Serialize() map[string]any {
	result := e.Attached.Serialize()
	result["Lifetime"] = e.Lifetime
	result["FadeTime"] = e.FadeTime
	if e.DeleteEntityOnExpiry {
		result["DeleteEntityOnExpiry"] = e.DeleteEntityOnExpiry
	}
	return result
}
