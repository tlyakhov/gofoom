// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"tlyakhov/gofoom/ecs"

	"github.com/spf13/cast"
)

type Spawner struct {
	ecs.Attached `editable:"^"`

	DeleteSpawnedOnDetach bool `editable:"Delete Spawned on Detach"`

	// See behaviors.Spawnee for the other side of this relation
	Spawned map[ecs.Entity]int64
}

func (s *Spawner) MultiAttachable() bool {
	return true
}

func (s *Spawner) String() string {
	return "Spawner"
}

func (s *Spawner) OnDetach(e ecs.Entity) {
	defer s.Attached.OnDetach(e)
	if !s.IsAttached() || !s.DeleteSpawnedOnDetach {
		return
	}
	for e := range s.Spawned {
		ecs.Delete(e)
	}
}

func (s *Spawner) Construct(data map[string]any) {
	s.Attached.Construct(data)
	s.Spawned = make(map[ecs.Entity]int64)
	s.DeleteSpawnedOnDetach = true

	if data == nil {
		return
	}

	if v, ok := data["DeleteSpawnedOnDetach"]; ok {
		s.DeleteSpawnedOnDetach = cast.ToBool(v)
	}

	if v, ok := data["Spawned"]; ok {
		spawned := ecs.ParseEntityTable(v, false)
		for _, e := range spawned {
			if e != 0 {
				s.Spawned[e] = ecs.Simulation.SimTimestamp
			}
		}
	}
}

func (s *Spawner) Serialize() map[string]any {
	result := s.Attached.Serialize()

	if !s.DeleteSpawnedOnDetach {
		result["DeleteSpawnedOnDetach"] = false
	}
	if len(s.Spawned) > 0 {
		spawned := make(ecs.EntityTable, 0)
		for e := range s.Spawned {
			spawned.Set(e)
		}
		result["Spawned"] = spawned.Serialize()
	}

	return result
}
