// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package core

import (
	"maps"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/ecs"
)

type Selection struct {
	Exact     map[uint64]*Selectable
	Grouped   map[uint64]*Selectable
	Positions map[uint64]*concepts.Vector3
}

func NewSelection() *Selection {
	s := &Selection{}
	s.Clear()
	return s
}

func NewSelectionClone(toCopy *Selection) *Selection {
	s := &Selection{}
	s.Exact = maps.Clone(toCopy.Exact)
	s.Grouped = maps.Clone(toCopy.Grouped)
	s.Positions = maps.Clone(toCopy.Positions)
	return s
}

func (s *Selection) Clear() {
	s.Exact = make(map[uint64]*Selectable)
	s.Grouped = make(map[uint64]*Selectable)
	s.Positions = make(map[uint64]*concepts.Vector3)
}

func (s *Selection) Empty() bool {
	return len(s.Exact) == 0
}

func (s *Selection) Add(list ...*Selectable) {
	for _, item := range list {
		s.Exact[item.Hash()] = item
		s.Grouped[item.GroupHash()] = item
	}
}

func (s *Selection) Contains(item *Selectable) bool {
	return s.Exact[item.Hash()] != nil
}
func (s *Selection) ContainsGrouped(item *Selectable) bool {
	return s.Grouped[item.GroupHash()] != nil
}

func (sel *Selection) Normalize() {
	// Stores # of segments selected for each sector, or -1 if the sector itself
	// is part of the selection.
	segmentsForSector := make(map[ecs.Entity]int)

	// Count segments for sectors
	for _, s := range sel.Exact {
		if s.SectorSegment != nil {
			if num, ok := segmentsForSector[s.Sector.Entity]; ok {
				if num == -1 {
					continue
				}
				segmentsForSector[s.Sector.Entity]++
			} else {
				segmentsForSector[s.Sector.Entity] = 1
			}
		}
		if typeGroups[s.Type] == SelectableSector {
			segmentsForSector[s.Sector.Entity] = -1
		}
	}

	// Remove segments if the sector is selected
	for _, s := range sel.Exact {
		if s.SectorSegment != nil && segmentsForSector[s.Sector.Entity] == -1 {
			delete(sel.Exact, s.Hash())
			delete(sel.Grouped, s.GroupHash())
		} else if s.SectorSegment != nil && segmentsForSector[s.Sector.Entity] == len(s.Sector.Segments) {
			delete(sel.Exact, s.Hash())
			delete(sel.Grouped, s.GroupHash())
			sel.Add(SelectableFromSector(s.Sector))
			segmentsForSector[s.Sector.Entity] = -1
		}
	}
}

func (sel *Selection) SavePositions() {
	for _, s := range sel.Exact {
		switch s.Type {
		case SelectableSector:
			for _, seg := range s.Sector.Segments {
				// 4 bits for type, 16 bits for segment index, 44 bits for entity
				hash := (uint64(s.Type) << 60) | uint64(seg.Index<<44) | uint64(s.Entity)
				sel.Positions[hash] = &concepts.Vector3{seg.P[0], seg.P[1], 0}
			}
		case SelectableLow:
			fallthrough
		case SelectableMid:
			fallthrough
		case SelectableHi:
			fallthrough
		case SelectableSectorSegment:
			sel.Positions[s.Hash()] = &concepts.Vector3{s.SectorSegment.P[0], s.SectorSegment.P[1]}
		case SelectableActionWaypoint:
			sel.Positions[s.Hash()] = s.ActionWaypoint.P.Clone()
		case SelectableBody:
			sel.Positions[s.Hash()] = s.Body.Pos.Original.Clone()
		case SelectableInternalSegmentA:
			sel.Positions[s.Hash()] = &concepts.Vector3{s.InternalSegment.A[0], s.InternalSegment.A[1]}
		case SelectableInternalSegmentB:
			sel.Positions[s.Hash()] = &concepts.Vector3{s.InternalSegment.B[0], s.InternalSegment.B[1]}
		case SelectableInternalSegment:
			// 4 bits for type, 1 bit for A/B, 59 bits for entity
			hash := (uint64(s.Type) << 60) | uint64(s.Entity)
			sel.Positions[hash] = &concepts.Vector3{s.InternalSegment.A[0], s.InternalSegment.A[1]}
			hash = (uint64(s.Type) << 60) | uint64(1<<59) | uint64(s.Entity)
			sel.Positions[hash] = &concepts.Vector3{s.InternalSegment.B[0], s.InternalSegment.B[1]}
		}
	}
}

func (sel *Selection) LoadPositions() {
	for _, s := range sel.Exact {
		switch s.Type {
		case SelectableSector:
			for _, seg := range s.Sector.Segments {
				// 4 bits for type, 16 bits for segment index, 44 bits for entity
				hash := (uint64(s.Type) << 60) | uint64(seg.Index<<44) | uint64(s.Entity)
				seg.P.From(sel.Positions[hash].To2D())
			}
		case SelectableLow:
			fallthrough
		case SelectableMid:
			fallthrough
		case SelectableHi:
			fallthrough
		case SelectableSectorSegment:
			s.SectorSegment.P.From(sel.Positions[s.Hash()].To2D())
		case SelectableActionWaypoint:
			s.ActionWaypoint.P.From(sel.Positions[s.Hash()])
		case SelectableBody:
			s.Body.Pos.Original.From(sel.Positions[s.Hash()])
		case SelectableInternalSegmentA:
			s.InternalSegment.A.From(sel.Positions[s.Hash()].To2D())
		case SelectableInternalSegmentB:
			s.InternalSegment.B.From(sel.Positions[s.Hash()].To2D())
		case SelectableInternalSegment:
			// 4 bits for type, 1 bit for A/B, 59 bits for entity
			hash := (uint64(s.Type) << 60) | uint64(s.Entity)
			s.InternalSegment.A.From(sel.Positions[hash].To2D())
			hash = (uint64(s.Type) << 60) | uint64(1<<59) | uint64(s.Entity)
			s.InternalSegment.B.From(sel.Positions[hash].To2D())
		}
	}
}
