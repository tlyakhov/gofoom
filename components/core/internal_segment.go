// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package core

import (
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/ecs"

	"github.com/spf13/cast"
)

// InternalSegments must be sorted back-to-front alongside entities
type InternalSegment struct {
	ecs.Attached `editable:"^"`
	Segment      `editable:"^"`

	Bottom   float64 `editable:"Bottom"`
	Top      float64 `editable:"Top"`
	TwoSided bool    `editable:"Two sided?"`
}

var InternalSegmentCID ecs.ComponentID

func init() {
	InternalSegmentCID = ecs.RegisterComponent(&ecs.Column[InternalSegment, *InternalSegment]{Getter: GetInternalSegment})
}

func (x *InternalSegment) ComponentID() ecs.ComponentID {
	return InternalSegmentCID
}
func GetInternalSegment(u *ecs.Universe, e ecs.Entity) *InternalSegment {
	if asserted, ok := u.Component(e, InternalSegmentCID).(*InternalSegment); ok {
		return asserted
	}
	return nil
}

func (s *InternalSegment) String() string {
	return "Segment: (" + s.Segment.A.StringHuman() + ")-(" + s.Segment.B.StringHuman() + ")"
}

func (s *InternalSegment) DetachFromSectors() {
	col := ecs.ColumnFor[Sector](s.Universe, SectorCID)
	for i := range col.Cap() {
		if sector := col.Value(i); sector != nil {
			delete(sector.InternalSegments, s.Entity)
		}
	}
}

func (s *InternalSegment) AttachToSectors() {
	min := &concepts.Vector2{s.A[0], s.A[1]}
	max := &concepts.Vector2{s.B[0], s.B[1]}
	if min[0] > max[0] {
		min[0], max[0] = max[0], min[0]
	}
	if min[1] > max[1] {
		min[1], max[1] = max[1], min[1]
	}
	col := ecs.ColumnFor[Sector](s.Universe, SectorCID)
	for i := range col.Cap() {
		sector := col.Value(i)
		if sector == nil {
			continue
		}
		// This is missing the spanning case, where an internal segment is
		// passing through a sector, but neither endpoint is inside of it.
		// Seems like an edge case we don't really need to handle.
		if !sector.AABBIntersect2D(min, max, true) {
			delete(sector.InternalSegments, s.Entity)
			continue
		}
		if !sector.IsPointInside2D(s.A) &&
			!sector.IsPointInside2D(s.B) {
			delete(sector.InternalSegments, s.Entity)
			continue
		}

		sector.InternalSegments[s.Entity] = s
	}
}

func (s *InternalSegment) Construct(data map[string]any) {
	s.Attached.Construct(data)
	s.Segment.Construct(s.Universe, data)
	s.Bottom = 0
	s.Top = 64
	s.TwoSided = false

	if data == nil {
		return
	}

	if v, ok := data["Top"]; ok {
		s.Top = cast.ToFloat64(v)
	}
	if v, ok := data["Bottom"]; ok {
		s.Bottom = cast.ToFloat64(v)
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

	result["Top"] = s.Top
	result["Bottom"] = s.Bottom

	if s.TwoSided {
		result["TwoSided"] = true
	}
	return result
}
