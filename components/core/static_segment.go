package core

import (
	"tlyakhov/gofoom/concepts"
)

type StaticSegment struct {
	concepts.Attached `editable:"^"`
	Segment

	SectorEntityRef *concepts.EntityRef
}

var StaticSegmentComponentIndex int

func init() {
	StaticSegmentComponentIndex = concepts.DbTypes().Register(StaticSegment{}, SectorFromDb)
}

func StaticSegmentFromDb(entity *concepts.EntityRef) *StaticSegment {
	if asserted, ok := entity.Component(StaticSegmentComponentIndex).(*StaticSegment); ok {
		return asserted
	}
	return nil
}

func (s *StaticSegment) Construct(data map[string]any) {
	s.Attached.Construct(data)
	s.Segment.Construct(s.DB, data)

	if data == nil {
		return
	}
}

func (s *StaticSegment) Serialize() map[string]any {
	result := s.Attached.Serialize()
	seg := s.Segment.Serialize(false)
	for k, v := range seg {
		result[k] = v
	}
	return result
}
