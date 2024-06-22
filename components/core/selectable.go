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
	SelectableLo
	SelectableMid
	SelectableInternalSegment
	SelectableInternalSegmentA
	SelectableInternalSegmentB
	SelectableBody
)

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

func SelectableFromSegment(s *SectorSegment) *Selectable {
	return &Selectable{
		Type:          SelectableSectorSegment,
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
	for i, test := range list {
		if test.Type != target.Type {
			continue
		}
		switch test.Type {
		case SelectableSector:
			if target.Sector != test.Sector {
				continue
			}
		case SelectableSectorSegment:
			if target.SectorSegment != test.SectorSegment {
				continue
			}
		case SelectableBody:
			if target.Body != test.Body {
				continue
			}
		case SelectableInternalSegment:
			fallthrough
		case SelectableInternalSegmentA:
			fallthrough
		case SelectableInternalSegmentB:
			if target.InternalSegment != test.InternalSegment {
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
