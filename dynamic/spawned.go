// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package dynamic

import (
	"log"
	"strconv"
	"tlyakhov/gofoom/concepts"

	"github.com/spf13/cast"
)

type Spawned[T DynamicType] struct {
	Attached bool
	// Warning: the location within the struct matters for fast access. Do not
	// move or change order without updating Dynamic.Value
	Spawn T `editable:"Spawn"`
	Now   T
}

func (s *Spawned[T]) ResetToSpawn() {
	s.Now = s.Spawn
}

func (s *Spawned[T]) SetAll(v T) {
	s.Spawn = v
	s.ResetToSpawn()
}

func (s *Spawned[T]) Attach(sim *Simulation) {
	sim.Spawnables.Store(s, struct{}{})
	s.Attached = true
}

func (s *Spawned[T]) Detach(sim *Simulation) {
	sim.Spawnables.Delete(s)
	s.Attached = false
}

func (s *Spawned[T]) Serialize() map[string]any {
	result := make(map[string]any)

	switch sc := any(s).(type) {
	case *Spawned[int]:
		result["Spawn"] = strconv.Itoa(sc.Spawn)
		result["Now"] = strconv.Itoa(sc.Now)
	case *Spawned[float64]:
		result["Spawn"] = sc.Spawn
		result["Now"] = sc.Now
	case *Spawned[concepts.Vector2]:
		result["Spawn"] = sc.Spawn.Serialize()
		result["Now"] = sc.Now.Serialize()
	case *Spawned[concepts.Vector3]:
		result["Spawn"] = sc.Spawn.Serialize()
		result["Now"] = sc.Now.Serialize()
	case *Spawned[concepts.Vector4]:
		result["Spawn"] = sc.Spawn.Serialize(false)
		result["Now"] = sc.Now.Serialize(false)
	case *Spawned[concepts.Matrix2]:
		result["Spawn"] = sc.Spawn.Serialize()
		result["Now"] = sc.Now.Serialize()
	default:
		log.Panicf("Tried to serialize Spawned[T] %v where T has no serializer", s)
	}

	if result["Spawn"] == result["Now"] {
		delete(result, "Now")
	}

	return result
}

func (s *Spawned[T]) Construct(data map[string]any) {
	switch sc := any(s).(type) {
	case *Spawned[concepts.Matrix2]:
		sc.Spawn.SetIdentity()
	}

	if data == nil {
		s.ResetToSpawn()
		return
	}

	if v, ok := data["Spawn"]; ok {
		switch sc := any(s).(type) {
		case *Spawned[int]:
			sc.Spawn = cast.ToInt(v)
		case *Spawned[float64]:
			sc.Spawn = cast.ToFloat64(v)
		case *Spawned[concepts.Vector2]:
			sc.Spawn.Deserialize(v.(string))
		case *Spawned[concepts.Vector3]:
			sc.Spawn.Deserialize(v.(string))
		case *Spawned[concepts.Vector4]:
			sc.Spawn.Deserialize(v.(string))
		case *Spawned[concepts.Matrix2]:
			sc.Spawn.Deserialize(v.(string))
		default:
			log.Panicf("Tried to deserialize Spawned[T] %v where T has no serializer", s)
		}
	}
	if v, ok := data["Now"]; ok {
		switch sc := any(s).(type) {
		case *Spawned[int]:
			sc.Now = cast.ToInt(v)
		case *Spawned[float64]:
			sc.Now = cast.ToFloat64(v)
		case *Spawned[concepts.Vector2]:
			sc.Now.Deserialize(v.(string))
		case *Spawned[concepts.Vector3]:
			sc.Now.Deserialize(v.(string))
		case *Spawned[concepts.Vector4]:
			sc.Now.Deserialize(v.(string))
		case *Spawned[concepts.Matrix2]:
			sc.Now.Deserialize(v.(string))
		default:
			log.Panicf("Tried to deserialize Spawned[T] %v where T has no serializer", s)
		}
	} else {
		s.ResetToSpawn()
	}
}
