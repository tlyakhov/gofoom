// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package core

import (
	"tlyakhov/gofoom/concepts"
)

// InternalSegments must be:
// 1. Inside a sector
// 2. Not a portal
// 3. Sorted back-to-front alongside entities
type InternalSegment struct {
	concepts.Attached `editable:"^"`
	Segment           `editable:"^"`

	TwoSided        bool `editable:"Two sided?"`
	SectorEntityRef *concepts.EntityRef
}

var InternalSegmentComponentIndex int

func init() {
	InternalSegmentComponentIndex = concepts.DbTypes().Register(InternalSegment{}, SectorFromDb)
}

func InternalSegmentFromDb(entity *concepts.EntityRef) *InternalSegment {
	if asserted, ok := entity.Component(InternalSegmentComponentIndex).(*InternalSegment); ok {
		return asserted
	}
	return nil
}

func (s *InternalSegment) Sector() *Sector {
	return SectorFromDb(s.SectorEntityRef)
}

func (s *InternalSegment) String() string {
	return "Segment: (" + s.Segment.A.StringHuman() + ")-(" + s.Segment.B.StringHuman() + ")"
}

func (s *InternalSegment) Construct(data map[string]any) {
	s.Attached.Construct(data)
	s.Segment.Construct(s.DB, data)

	s.TwoSided = false

	if data == nil {
		return
	}

	if v, ok := data["TwoSided"]; ok {
		s.TwoSided = v.(bool)
	}
}

func (s *InternalSegment) Serialize() map[string]any {
	result := s.Attached.Serialize()
	seg := s.Segment.Serialize(true)
	for k, v := range seg {
		result[k] = v
	}

	if s.TwoSided {
		result["TwoSided"] = true
	}
	return result
}
