// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"tlyakhov/gofoom/containers"
	"tlyakhov/gofoom/ecs"
)

type ParticleEmitter struct {
	ecs.Attached `editable:"^"`

	Particles containers.Set[ecs.Entity]
	Spawned   map[ecs.Entity]int64
}

var ParticleEmitterCID ecs.ComponentID

func init() {
	ParticleEmitterCID = ecs.RegisterComponent(&ecs.Column[ParticleEmitter, *ParticleEmitter]{Getter: GetParticleEmitter}, "")
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

	pe.Particles = make(containers.Set[ecs.Entity])
	pe.Spawned = make(map[ecs.Entity]int64)

	if data == nil {
		return
	}

	if v, ok := data["Particles"]; ok {
		pe.Particles = ecs.DeserializeEntities(v.([]any))
		for e := range pe.Particles {
			pe.Spawned[e] = pe.ECS.Timestamp
		}
	}
}

func (pe *ParticleEmitter) Serialize() map[string]any {
	result := pe.Attached.Serialize()

	if len(pe.Particles) > 0 {
		result["Particles"] = ecs.SerializeEntities(pe.Particles)
	}

	return result
}
