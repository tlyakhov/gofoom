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

// TODO: Paths should probably be refactored to be scriptable Actions instead.
// An action could be to move to a point sure, but it could also be to change
// angle, fire a weapon, etc...
type Path struct {
	ecs.Attached `editable:"^"`

	Segments []*PathSegment `editable:"Segments"`
}

var PathComponentIndex int

func init() {
	PathComponentIndex = ecs.RegisterComponent(&ecs.Column[Path, *Path]{Getter: GetPath})
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

func (p *Path) Recalculate() {
	filtered := make([]*PathSegment, 0)
	var prev *PathSegment
	for i, segment := range p.Segments {
		segment.Index = len(filtered)
		filtered = append(filtered, segment)
		segment.Path = p

		nextIndex := (i + 1) % len(p.Segments)
		next := p.Segments[nextIndex]
		/*if nextIndex == 0 && !p.Loop {
			segment.Length = 0
			segment.Next = nil
			next.Prev = nil
			continue
		}*/

		// Filter out degenerate segments.
		if prev != nil && prev.P == segment.P {
			prev.Next = next
			next.Prev = prev
			continue
		}
		segment.Next = next
		next.Prev = segment
		prev = segment
		segment.Length = next.P.Dist(&segment.P)
	}
	p.Segments = filtered
}
