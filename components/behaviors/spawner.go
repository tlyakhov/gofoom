// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"tlyakhov/gofoom/ecs"
)

type Spawner struct {
	ecs.Attached `editable:"^"`

	Spawned map[ecs.Entity]int64
}

func (s *Spawner) MultiAttachable() bool {
	return true
}

func (s *Spawner) String() string {
	return "Spawner"
}

func (s *Spawner) Construct(data map[string]any) {
	s.Attached.Construct(data)
	s.Spawned = make(map[ecs.Entity]int64)

	if data == nil {
		return
	}

	if v, ok := data["Spawned"]; ok {
		spawned := ecs.ParseEntityTable(v, false)
		for _, e := range spawned {
			if e != 0 {
				s.Spawned[e] = ecs.Simulation.Timestamp
			}
		}
	}
}

func (s *Spawner) Serialize() map[string]any {
	result := s.Attached.Serialize()

	if len(s.Spawned) > 0 {
		spawned := make(ecs.EntityTable, 0)
		for e := range s.Spawned {
			spawned.Set(e)
		}
		result["Spawned"] = spawned.Serialize()
	}

	return result
}
