// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"fmt"
	"tlyakhov/gofoom/ecs"

	"github.com/spf13/cast"
)

// Keeps track of which entity spawned this one.
type Spawnee struct {
	ecs.Attached `editable:"^"`

	Spawner ecs.Entity // behaviors.Spawner
}

func (s *Spawnee) String() string {
	return fmt.Sprintf("Spawned from %v", s.Spawner.String())
}

func (s *Spawnee) OnDetach(e ecs.Entity) {
	defer s.Attached.OnDetach(e)
	if !s.IsAttached() || s.Spawner == 0 {
		return
	}
	if spawner := GetSpawner(s.Spawner); spawner != nil {
		delete(spawner.Spawned, e)
	}
}

func (s *Spawnee) Construct(data map[string]any) {
	s.Attached.Construct(data)

	if data == nil {
		return
	}

	if v, ok := data["Spawner"]; ok {
		s.Spawner, _ = ecs.ParseEntity(cast.ToString(v))
	}
}

func (s *Spawnee) Serialize() map[string]any {
	result := s.Attached.Serialize()

	if s.Spawner != 0 {
		result["Spawner"] = s.Spawner.Serialize()
	}

	return result
}
