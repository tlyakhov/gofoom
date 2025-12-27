// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package core

import (
	"fmt"
	"math"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/dynamic"
	"tlyakhov/gofoom/ecs"

	"github.com/spf13/cast"
)

type SectorPlane struct {
	*Sector

	Z       dynamic.DynamicValue[float64] `editable:"Z"`
	Normal  concepts.Vector3              `editable:"Normal" edit_type:"Normal"`
	Target  ecs.Entity                    `editable:"Target" edit_type:"Sector"`
	Surface materials.Surface             `editable:"Surf"`
	Scripts []*Script                     `editable:"Scripts"`

	// This is only valid for inner sectors
	Ignore bool `editable:"Ignore"`

	XYDet float64
}

func (s *SectorPlane) String() string {
	return fmt.Sprintf("Plane (Z: %v, Normal: %v)", s.Z.Now, s.Normal)
}

func (s *SectorPlane) Construct(data map[string]any) {
	s.Surface.Construct(data)
	s.Z.Construct(nil)

	if data == nil {
		return
	}

	if v, ok := data["Z"]; ok {
		s.Z.Construct(v)
	}
	if v, ok := data["Normal"]; ok {
		s.Normal.Deserialize(v.(string))
	}
	if v, ok := data["Surface"]; ok {
		s.Surface.Construct(v.(map[string]any))
	}
	if v, ok := data["Target"]; ok {
		s.Target, _ = ecs.ParseEntity(v.(string))
	}
	if v, ok := data["Scripts"]; ok {
		s.Scripts = ecs.ConstructSlice[*Script](v, nil)
	}
	if v, ok := data["Ignore"]; ok {
		s.Ignore = cast.ToBool(v)
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
		result["Target"] = s.Target.Serialize()
	}
	if len(s.Scripts) > 0 {
		result["Scripts"] = ecs.SerializeSlice(s.Scripts)
	}

	if s.Ignore {
		result["Ignore"] = true
	}

	return result
}

func (s *SectorPlane) Precompute() {
	if s.Sector == nil || len(s.Sector.Segments) == 0 {
		s.XYDet = 0
		return
	}

	s.XYDet = s.Normal[1]*s.Segments[0].P.Render[1] +
		s.Normal[0]*s.Segments[0].P.Render[0]
}

func (s *SectorPlane) ZAt(isect *concepts.Vector2) float64 {
	det := s.Normal[2]*s.Z.Render + s.XYDet

	return (det - s.Normal[0]*isect[0] - s.Normal[1]*isect[1]) / s.Normal[2]
}
