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
	Spawn    T `editable:"Spawn"`
	Now      T
}

func (s *Spawned[T]) ResetToSpawn() {
	s.Now = s.Spawn
}

func (s *Spawned[T]) SetAll(v T) {
	s.Spawn = v
	s.ResetToSpawn()
}

func (s *Spawned[T]) Attach(sim *Simulation) {
	sim.Spawnables[s] = struct{}{}
	s.Attached = true
}

func (s *Spawned[T]) Detach(sim *Simulation) {
	delete(sim.Spawnables, s)
	s.Attached = false
}

func (s *Spawned[T]) serializeValue(v T) any {
	switch typed := any(v).(type) {
	case int:
		return strconv.Itoa(typed)
	case float64:
		return typed
	case concepts.Vector2:
		return typed.Serialize()
	case concepts.Vector3:
		return typed.Serialize()
	case concepts.Vector4:
		return typed.Serialize(false)
	case concepts.Matrix2:
		return typed.Serialize()
	default:
		log.Panicf("Tried to serialize Spawned[T] %v where T has no serializer", s)
	}
	return nil
}

func (s *Spawned[T]) deserializeValue(v any) any {
	var r T
	switch typed := any(r).(type) {
	case int:
		return cast.ToInt(v)
	case float64:
		return cast.ToFloat64(v)
	case concepts.Vector2:
		typed.Deserialize(v.(string))
		return typed
	case concepts.Vector3:
		typed.Deserialize(v.(string))
		return typed
	case concepts.Vector4:
		typed.Deserialize(v.(string))
		return typed
	case concepts.Matrix2:
		typed.Deserialize(v.(string))
		return typed
	default:
		log.Panicf("Tried to deserialize Spawned[T] %v where T has no serializer", s)
	}
	return r
}

func (s *Spawned[T]) Serialize() any {
	if s.Spawn == s.Now {
		return s.serializeValue(s.Spawn)
	}

	result := make(map[string]any)
	result["Spawn"] = s.serializeValue(s.Spawn)
	result["Now"] = s.serializeValue(s.Now)

	return result
}

func (s *Spawned[T]) Construct(data any) {
	switch sc := any(s).(type) {
	case *Spawned[concepts.Matrix2]:
		sc.Spawn.SetIdentity()
	}

	if data == nil {
		s.ResetToSpawn()
		return
	}
	var params map[string]any
	var ok bool
	if params, ok = data.(map[string]any); !ok || params["Spawn"] == nil {
		s.Spawn = s.deserializeValue(data).(T)
		s.ResetToSpawn()
		return
	}

	if v, ok := params["Spawn"]; ok {
		s.Spawn = s.deserializeValue(v).(T)
	}
	if v, ok := params["Now"]; ok {
		s.Now = s.deserializeValue(v).(T)
	} else {
		s.ResetToSpawn()
	}
}
