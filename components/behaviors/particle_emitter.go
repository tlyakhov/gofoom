// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"strconv"
	"tlyakhov/gofoom/ecs"

	"github.com/spf13/cast"
)

type ParticleEmitter struct {
	ecs.Attached `editable:"^"`

	Lifetime float64 `editable:"Lifetime"`  // ms
	FadeTime float64 `editable:"Fade Time"` // ms
	Limit    int     `editable:"Particle Count Limit"`
	XYSpread float64 `editable:"XY Spread"` // Degrees
	ZSpread  float64 `editable:"Z Spread"`  // Degrees
	Vel      float64 `editable:"Velocity"`

	Spawner ecs.Entity `editable:"Spawner" edit_type:"Spawner"`
}

func (pe *ParticleEmitter) String() string {
	return "Particle Emitter"
}

func (pe *ParticleEmitter) Construct(data map[string]any) {
	pe.Attached.Construct(data)

	pe.Lifetime = 5000
	pe.FadeTime = 1000
	pe.Limit = 30
	pe.XYSpread = 10
	pe.ZSpread = 10
	pe.Vel = 15
	pe.Spawner = 0

	if data == nil {
		return
	}

	if v, ok := data["Lifetime"]; ok {
		pe.Lifetime = cast.ToFloat64(v)
	}
	if v, ok := data["FadeTime"]; ok {
		pe.FadeTime = cast.ToFloat64(v)
	}
	if v, ok := data["XYSpread"]; ok {
		pe.XYSpread = cast.ToFloat64(v)
	}
	if v, ok := data["ZSpread"]; ok {
		pe.ZSpread = cast.ToFloat64(v)
	}
	if v, ok := data["Vel"]; ok {
		pe.Vel = cast.ToFloat64(v)
	}
	if v, ok := data["Limit"]; ok {
		pe.Limit = cast.ToInt(v)
	}
	if v, ok := data["Spawner"]; ok {
		pe.Spawner, _ = ecs.ParseEntity(v.(string))
	}
}

func (pe *ParticleEmitter) Serialize() map[string]any {
	result := pe.Attached.Serialize()

	if pe.Lifetime != 5000 {
		result["Lifetime"] = pe.Lifetime
	}
	if pe.FadeTime != 1000 {
		result["FadeTime"] = pe.FadeTime
	}
	if pe.Limit != 100 {
		result["Limit"] = strconv.Itoa(pe.Limit)
	}
	if pe.Spawner != 0 {
		result["Spawner"] = pe.Spawner.Serialize()
	}
	result["XYSpread"] = pe.XYSpread
	result["ZSpread"] = pe.ZSpread
	result["Vel"] = pe.Vel

	return result
}
