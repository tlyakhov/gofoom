// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package core

import (
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/ecs"
)

type PathSegment struct {
	ECS *ecs.ECS

	P concepts.Vector3 `editable:"Position"`

	// Pre-calculated attributes
	Index  int
	Length float64
	Path   *Path
	Next   *PathSegment
	Prev   *PathSegment
}

func (s *PathSegment) Construct(db *ecs.ECS, data map[string]any) {
	s.ECS = db
	s.P = concepts.Vector3{}

	if data == nil {
		return
	}

	s.P.Deserialize(data)
}

func (s *PathSegment) Serialize() map[string]any {
	result := make(map[string]any)
	result["X"] = s.P[0]
	result["Y"] = s.P[1]
	result["Z"] = s.P[2]

	return result
}
