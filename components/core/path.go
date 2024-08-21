// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package core

import (
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/ecs"
)

// This component represents a path. Examples:
// - Animated entities and NPCs can follow this
// - Special FX can be shaped by paths
// TODO: Use https://www.youtube.com/watch?v=KPoeNZZ6H4s
type Path struct {
	ecs.Attached `editable:"^"`

	Segments []*PathSegment
}

var PathComponentIndex int

func init() {
	PathComponentIndex = ecs.Types().Register(Path{}, GetPath)
}

func GetPath(db *ecs.ECS, e ecs.Entity) *Path {
	if asserted, ok := db.Component(e, PathComponentIndex).(*Path); ok {
		return asserted
	}
	return nil
}

func (p *Path) String() string {
	return "Path: " + p.Entity.String()
}

func (p *Path) AddSegment(x float64, y float64, z float64) *PathSegment {
	segment := new(PathSegment)
	segment.Construct(p.ECS, nil)
	segment.Path = p
	segment.P = concepts.Vector3{x, y, z}
	p.Segments = append(p.Segments, segment)
	return segment
}

func (p *Path) Construct(data map[string]any) {
	p.Attached.Construct(data)

	p.Segments = make([]*PathSegment, 0)

	if data == nil {
		return
	}

	if v, ok := data["Segments"]; ok {
		jsonSegments := v.([]any)
		p.Segments = make([]*PathSegment, len(jsonSegments))
		for i, jsonSegment := range jsonSegments {
			segment := new(PathSegment)
			segment.Path = p
			segment.Construct(p.ECS, jsonSegment.(map[string]any))
			p.Segments[i] = segment
		}
	}

	p.Recalculate()
}

func (p *Path) Serialize() map[string]any {
	result := p.Attached.Serialize()

	segments := []any{}
	for _, seg := range p.Segments {
		segments = append(segments, seg.Serialize())
	}
	result["Segments"] = segments
	return result
}

func (s *Path) Recalculate() {
	filtered := make([]*PathSegment, 0)
	var prev *PathSegment
	for i, segment := range s.Segments {
		next := s.Segments[(i+1)%len(s.Segments)]
		// Filter out degenerate segments.
		if prev != nil && prev.P == segment.P {
			prev.Next = next
			next.Prev = prev
			continue
		}
		filtered = append(filtered, segment)
		segment.Next = next
		next.Prev = segment
		prev = segment
		segment.Path = s
		segment.Recalculate()
	}
	s.Segments = filtered
}
