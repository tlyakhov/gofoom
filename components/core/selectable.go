// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package core

import (
	"tlyakhov/gofoom/concepts"
)

//go:generate go run github.com/dmarkham/enumer -type=SelectableType -json
type SelectableType int

const (
	SelectableEntityRef SelectableType = iota
	SelectableSector
	SelectableSectorSegment
	SelectableCeiling
	SelectableFloor
	SelectableHi
	SelectableLow
	SelectableMid
	SelectableInternalSegment
	SelectableInternalSegmentA
	SelectableInternalSegmentB
	SelectableBody
)

var typeGroups = map[SelectableType]SelectableType{
	// Sectors
	SelectableSector:  SelectableSector,
	SelectableCeiling: SelectableSector,
	SelectableFloor:   SelectableSector,
	// Segments
	SelectableSectorSegment: SelectableSectorSegment,
	SelectableHi:            SelectableSectorSegment,
	SelectableLow:           SelectableSectorSegment,
	SelectableMid:           SelectableSectorSegment,
	// Internal Segments
	SelectableInternalSegment:  SelectableInternalSegment,
	SelectableInternalSegmentA: SelectableInternalSegment,
	SelectableInternalSegmentB: SelectableInternalSegment,
	// Other
	SelectableBody:      SelectableBody,
	SelectableEntityRef: SelectableEntityRef,
}

// Selectable represents something an editor or player can pick. Initially this
// was implemented via just passing `any` instances around, but that had the
// limitation of not being able to select parts of objects that weren't explicit
// types. (for example, only one point of an internal segment)
type Selectable struct {
	Type            SelectableType
	Sector          *Sector
	Body            *Body
	SectorSegment   *SectorSegment
	InternalSegment *InternalSegment
	Ref             *concepts.EntityRef
}

func SelectableFromSector(s *Sector) *Selectable {
	return &Selectable{Type: SelectableSector, Sector: s, Ref: s.EntityRef}
}

func SelectableFromFloor(s *Sector) *Selectable {
	return &Selectable{Type: SelectableFloor, Sector: s, Ref: s.EntityRef}
}

func SelectableFromCeil(s *Sector) *Selectable {
	return &Selectable{Type: SelectableCeiling, Sector: s, Ref: s.EntityRef}
}

func SelectableFromSegment(s *SectorSegment) *Selectable {
	return &Selectable{
		Type:          SelectableSectorSegment,
		Sector:        s.Sector,
		SectorSegment: s,
		Ref:           s.Sector.EntityRef}
}

func SelectableFromWall(s *SectorSegment, t SelectableType) *Selectable {
	return &Selectable{
		Type:          t,
		Sector:        s.Sector,
		SectorSegment: s,
		Ref:           s.Sector.EntityRef}
}

func SelectableFromBody(b *Body) *Selectable {
	return &Selectable{
		Type:   SelectableBody,
		Sector: b.Sector(),
		Body:   b,
		Ref:    b.EntityRef}
}

func SelectableFromInternalSegment(s *InternalSegment) *Selectable {
	return &Selectable{
		Type:            SelectableInternalSegment,
		Sector:          s.Sector(),
		InternalSegment: s,
		Ref:             s.EntityRef}
}

func SelectableFromInternalSegmentA(s *InternalSegment) *Selectable {
	return &Selectable{Type: SelectableInternalSegmentA,
		Sector:          s.Sector(),
		InternalSegment: s,
		Ref:             s.EntityRef}
}

func SelectableFromInternalSegmentB(s *InternalSegment) *Selectable {
	return &Selectable{Type: SelectableInternalSegmentB,
		Sector:          s.Sector(),
		InternalSegment: s,
		Ref:             s.EntityRef}
}

func SelectableFromEntityRef(ref *concepts.EntityRef) *Selectable {
	if sector := SectorFromDb(ref); sector != nil {
		return SelectableFromSector(sector)
	}
	if body := BodyFromDb(ref); body != nil {
		return SelectableFromBody(body)
	}
	if seg := InternalSegmentFromDb(ref); seg != nil {
		return SelectableFromInternalSegment(seg)
	}
	return &Selectable{Type: SelectableEntityRef, Ref: ref}
}

func (target *Selectable) IndexIn(list []*Selectable) int {
	return target.search(list, false)
}

func (target *Selectable) ExactIndexIn(list []*Selectable) int {
	return target.search(list, true)
}

func (target *Selectable) search(list []*Selectable, exact bool) int {
	for i, test := range list {
		if exact && test.Type != target.Type {
			continue
		} else if !exact && typeGroups[test.Type] != typeGroups[target.Type] {
			continue
		}

		switch test.Type {
		// Sector selectables
		case SelectableFloor:
			fallthrough
		case SelectableCeiling:
			fallthrough
		case SelectableSector:
			if target.Sector != test.Sector {
				continue
			}
		// Segment selectables
		case SelectableLow:
			fallthrough
		case SelectableMid:
			fallthrough
		case SelectableHi:
			fallthrough
		case SelectableSectorSegment:
			if target.SectorSegment != test.SectorSegment {
				continue
			}
		// InternalSegment selectables
		case SelectableInternalSegment:
			fallthrough
		case SelectableInternalSegmentA:
			fallthrough
		case SelectableInternalSegmentB:
			if target.InternalSegment != test.InternalSegment {
				continue
			}
		// Other
		case SelectableBody:
			if target.Body != test.Body {
				continue
			}
		case SelectableEntityRef:
			if target.Ref.Entity != test.Ref.Entity {
				continue
			}
		}
		return i
	}
	return -1
}
func (s *Selectable) AddToList(list *[]*Selectable) bool {
	if s.IndexIn(*list) >= 0 {
		return false
	}
	*list = append(*list, s)
	return false
}

// Serialize saves the data for whatever the selectable is holding, which may or
// may not be an Entity (could be a component of one)
func (s *Selectable) Serialize() any {
	return s.Ref.DB.SerializeEntity(s.Ref.Entity)
}

func (s *Selectable) Transform(m *concepts.Matrix2) {
	switch s.Type {
	case SelectableSector:
		for _, seg := range s.Sector.Segments {
			m.ProjectSelf(&seg.P)
		}
		s.Sector.Recalculate()
	case SelectableLow:
		fallthrough
	case SelectableMid:
		fallthrough
	case SelectableHi:
		fallthrough
	case SelectableSectorSegment:
		m.ProjectSelf(&s.SectorSegment.P)
		s.Sector.Recalculate()
	case SelectableBody:
		m.ProjectXYSelf(&s.Body.Pos.Original)
		s.Body.Pos.Reset()
	case SelectableInternalSegmentA:
		m.ProjectSelf(s.InternalSegment.A)
		s.InternalSegment.Recalculate()
	case SelectableInternalSegmentB:
		m.ProjectSelf(s.InternalSegment.B)
		s.InternalSegment.Recalculate()
	case SelectableInternalSegment:
		m.ProjectSelf(s.InternalSegment.A)
		m.ProjectSelf(s.InternalSegment.B)
		s.InternalSegment.Recalculate()
	}
}
