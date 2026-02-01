// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"log"
	"tlyakhov/gofoom/ecs"

	"github.com/spf13/cast"
)

type AutoSpawn int

//go:generate go run github.com/dmarkham/enumer -type=AutoSpawn -json
const (
	AutoSpawnNone AutoSpawn = iota
	AutoSpawnOnLoad
	AutoSpawnDeleteOnLoad
)

type Spawner struct {
	ecs.Attached `editable:"^"`

	DeleteSpawnedOnDetach bool            `editable:"Propagate deletion"`
	PreserveLinks         bool            `editable:"Preserve Links"`
	Auto                  AutoSpawn       `editable:"Behavior on load"`
	Targets               ecs.EntityTable `editable:"Targets"`

	// See behaviors.Spawnee for the other side of this relation
	Spawned map[ecs.Entity]int64
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
	s.PreserveLinks = false
	s.Auto = AutoSpawnOnLoad
	s.Targets = nil

	if data == nil {
		return
	}

	if v, ok := data["DeleteSpawnedOnDetach"]; ok {
		s.DeleteSpawnedOnDetach = cast.ToBool(v)
	}
	if v, ok := data["PreserveLinks"]; ok {
		s.PreserveLinks = cast.ToBool(v)
	}
	if v, ok := data["Auto"]; ok {
		if auto, err := AutoSpawnString(v.(string)); err == nil {
			s.Auto = auto
		} else {
			log.Printf("error parsing spawner auto %v", v)
		}
	}
	if v, ok := data["Targets"]; ok {
		s.Targets = ecs.ParseEntityTable(v, true)
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
	if s.PreserveLinks {
		result["PreserveLinks"] = true
	}
	if s.Auto != AutoSpawnOnLoad {
		result["Auto"] = s.Auto.String()
	}
	if len(s.Targets) != 0 {
		result["Targets"] = s.Targets.Serialize()
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
