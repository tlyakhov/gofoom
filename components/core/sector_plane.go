// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package core

import (
	"math"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/dynamic"
	"tlyakhov/gofoom/ecs"

	"github.com/spf13/cast"
)

type SectorPlane struct {
	ECS *ecs.ECS
	*Sector

	Z       dynamic.DynamicValue[float64] `editable:"Z"`
	Normal  concepts.Vector3              `editable:"Normal"`
	Target  ecs.Entity                    `editable:"Target" edit_type:"Sector"`
	Surface materials.Surface             `editable:"Surf"`
	Scripts []*Script                     `editable:"Scripts"`
}

func (s *SectorPlane) Construct(sector *Sector, data map[string]any) {
	s.Sector = sector
	s.Surface.Construct(sector.ECS, data)
	s.Z.Construct(nil)
	s.ECS = sector.ECS

	if data == nil {
		return
	}

	if v, ok := data["Z"]; ok {
		if v2, err := cast.ToFloat64E(v); err == nil {
			v = map[string]any{"Spawn": v2}
		}
		s.Z.Construct(v.(map[string]any))
	}
	if v, ok := data["Normal"]; ok {
		s.Normal.Deserialize(v.(string))
	}
	if v, ok := data["Surface"]; ok {
		s.Surface.Construct(s.ECS, v.(map[string]any))
	}
	if v, ok := data["Target"]; ok {
		s.Target, _ = ecs.ParseEntity(v.(string))
	}
	if v, ok := data["Scripts"]; ok {
		s.Scripts = ecs.ConstructSlice[*Script](s.ECS, v, nil)
	}
}

func (s *SectorPlane) Serialize() map[string]any {
	result := make(map[string]any)
	result["Z"] = s.Z.Serialize()
	result["Surface"] = s.Surface.Serialize()

	if s.Normal[0] != 0 || s.Normal[1] != 0 || math.Abs(s.Normal[2]) != 1 {
		result["Normal"] = s.Normal.Serialize()
	}
	if s.Target != 0 {
		result["Target"] = s.Target.String()
	}
	if len(s.Scripts) > 0 {
		result["Scripts"] = ecs.SerializeSlice(s.Scripts)
	}

	return result
}

func (s *SectorPlane) ZAt(stage dynamic.DynamicStage, isect *concepts.Vector2) float64 {
	z := s.Z.Value(stage)
	if s.Sector == nil || len(s.Sector.Segments) == 0 {
		return z
	}
	det := s.Normal[2]*z + s.Normal[1]*s.Segments[0].P[1] +
		s.Normal[0]*s.Segments[0].P[0]

	z = (det - s.Normal[0]*isect[0] - s.Normal[1]*isect[1]) / s.Normal[2]

	return z
}
