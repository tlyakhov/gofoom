// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"strconv"
	"tlyakhov/gofoom/ecs"
)

type ParticleEmitter struct {
	ecs.Attached `editable:"^"`

	Lifetime float64    `editable:"Lifetime"`  // ms
	FadeTime float64    `editable:"Fade Time"` // ms
	Limit    int        `editable:"Particle Count Limit"`
	Source   ecs.Entity `editable:"Source" edit_type:"Material"`
	XYSpread float64    `editable:"XY Spread"` // Degrees
	ZSpread  float64    `editable:"Z Spread"`  // Degrees
	Vel      float64    `editable:"Velocity"`

	Spawned map[ecs.Entity]int64
}

var ParticleEmitterCID ecs.ComponentID

func init() {
	ParticleEmitterCID = ecs.RegisterComponent(&ecs.Column[ParticleEmitter, *ParticleEmitter]{Getter: GetParticleEmitter})
}

func GetParticleEmitter(db *ecs.ECS, e ecs.Entity) *ParticleEmitter {
	if asserted, ok := db.Component(e, ParticleEmitterCID).(*ParticleEmitter); ok {
		return asserted
	}
	return nil
}

func (pe *ParticleEmitter) String() string {
	return "Particle Emitter"
}

func (pe *ParticleEmitter) Construct(data map[string]any) {
	pe.Attached.Construct(data)

	pe.Lifetime = 5000
	pe.FadeTime = 1000
	pe.Limit = 30
	pe.Source = 0
	pe.XYSpread = 10
	pe.ZSpread = 10
	pe.Vel = 15

	pe.Spawned = make(map[ecs.Entity]int64)

	if data == nil {
		return
	}

	if v, ok := data["Particles"]; ok {
		var particles ecs.EntityTable
		if s, ok := v.(string); ok {
			particles = ecs.ParseEntityCSV(s)
		} else {
			particles = ecs.DeserializeEntities(v.([]any))
		}
		for _, e := range particles {
			if e != 0 {
				pe.Spawned[e] = pe.ECS.Timestamp
			}
		}
	}

	if v, ok := data["Lifetime"]; ok {
		pe.Lifetime = v.(float64)
	}
	if v, ok := data["FadeTime"]; ok {
		pe.FadeTime = v.(float64)
	}
	if v, ok := data["XYSpread"]; ok {
		pe.XYSpread = v.(float64)
	}
	if v, ok := data["ZSpread"]; ok {
		pe.ZSpread = v.(float64)
	}
	if v, ok := data["Vel"]; ok {
		pe.Vel = v.(float64)
	}
	if v, ok := data["Limit"]; ok {
		pe.Limit, _ = strconv.Atoi(v.(string))
	}
	if v, ok := data["Source"]; ok {
		pe.Source, _ = ecs.ParseEntity(v.(string))
	}
}

func (pe *ParticleEmitter) Serialize() map[string]any {
	result := pe.Attached.Serialize()

	if len(pe.Spawned) > 0 {
		particles := make(ecs.EntityTable, 0)
		for e := range pe.Spawned {
			particles.Set(e)
		}
		result["Particles"] = particles.Serialize()
	}

	if pe.Lifetime != 5000 {
		result["Lifetime"] = pe.Lifetime
	}
	if pe.FadeTime != 1000 {
		result["FadeTime"] = pe.FadeTime
	}
	if pe.Limit != 100 {
		result["Limit"] = strconv.Itoa(pe.Limit)
	}
	if pe.Source != 0 {
		result["Source"] = pe.Source.String()
	}
	result["XYSpread"] = pe.XYSpread
	result["ZSpread"] = pe.ZSpread
	result["Vel"] = pe.Vel

	return result
}
