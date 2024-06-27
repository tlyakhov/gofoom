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
		InternalSegment: s,
		Ref:             s.EntityRef}
}

func SelectableFromInternalSegmentA(s *InternalSegment) *Selectable {
	return &Selectable{Type: SelectableInternalSegmentA,
		InternalSegment: s,
		Ref:             s.EntityRef}
}

func SelectableFromInternalSegmentB(s *InternalSegment) *Selectable {
	return &Selectable{Type: SelectableInternalSegmentB,
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
	if s.search(*list, true) >= 0 {
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

func (s *Selectable) SavePositions() []concepts.Vector3 {
	positions := make([]concepts.Vector3, 0, 1)

	switch s.Type {
	case SelectableSector:
		i := 0
		for _, seg := range s.Sector.Segments {
			positions = append(positions, concepts.Vector3{})
			seg.P.To3D(&positions[i])
			i++
		}
	case SelectableLow:
		fallthrough
	case SelectableMid:
		fallthrough
	case SelectableHi:
		fallthrough
	case SelectableSectorSegment:
		positions = append(positions, concepts.Vector3{})
		s.SectorSegment.P.To3D(&positions[0])
	case SelectableBody:
		positions = append(positions, s.Body.Pos.Original)
	case SelectableInternalSegmentA:
		positions = append(positions, concepts.Vector3{})
		s.InternalSegment.A.To3D(&positions[0])
	case SelectableInternalSegmentB:
		positions = append(positions, concepts.Vector3{})
		s.InternalSegment.B.To3D(&positions[0])
	case SelectableInternalSegment:
		positions = append(positions, concepts.Vector3{}, concepts.Vector3{})
		s.InternalSegment.A.To3D(&positions[0])
		s.InternalSegment.B.To3D(&positions[1])
	}
	return positions
}

func (s *Selectable) LoadPositions(positions []concepts.Vector3) int {
	switch s.Type {
	case SelectableSector:
		i := 0
		for _, seg := range s.Sector.Segments {
			seg.P.From(positions[i].To2D())
			i++
		}
		return i
	case SelectableLow:
		fallthrough
	case SelectableMid:
		fallthrough
	case SelectableHi:
		fallthrough
	case SelectableSectorSegment:
		s.SectorSegment.P.From(positions[0].To2D())
	case SelectableBody:
		s.Body.Pos.Original = positions[0]
	case SelectableInternalSegmentA:
		s.InternalSegment.A.From(positions[0].To2D())
	case SelectableInternalSegmentB:
		s.InternalSegment.B.From(positions[0].To2D())
	case SelectableInternalSegment:
		s.InternalSegment.A.From(positions[0].To2D())
		s.InternalSegment.B.From(positions[1].To2D())
		return 2
	}
	return 1
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

func (s *Selectable) Recalculate() {
	switch s.Type {
	case SelectableSector:
		fallthrough
	case SelectableLow:
		fallthrough
	case SelectableMid:
		fallthrough
	case SelectableHi:
		fallthrough
	case SelectableSectorSegment:
		s.Sector.Recalculate()
	case SelectableBody:
		s.Body.Pos.Reset()
	case SelectableInternalSegmentA:
		fallthrough
	case SelectableInternalSegmentB:
		fallthrough
	case SelectableInternalSegment:
		s.InternalSegment.Recalculate()
	}
}
