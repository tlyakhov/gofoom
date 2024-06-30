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

func (s *Selectable) Hash() uint64 {
	if s.SectorSegment != nil {
		// 4 bits for type, 16 bits for segment index, 44 bits for entity
		return (uint64(s.Type) << 60) | uint64(s.SectorSegment.Index<<44) | s.Ref.Entity
	}
	// 4 bits for type, 60 bits for entity
	return (uint64(s.Type) << 60) | s.Ref.Entity
}

func (s *Selectable) GroupHash() uint64 {
	if s.SectorSegment != nil {
		// 4 bits for type, 16 bits for segment index, 44 bits for entity
		return (uint64(SelectableSectorSegment) << 60) | uint64(s.SectorSegment.Index<<44) | s.Ref.Entity
	}
	// 4 bits for type, 60 bits for entity
	return (uint64(typeGroups[s.Type]) << 60) | s.Ref.Entity
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
