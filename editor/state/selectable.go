// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package state

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
)

type EditorSelectableType int

const (
	SelectableSector EditorSelectableType = iota
	SelectableSectorSegment
	SelectableBody
	SelectableInternalSegment
	SelectableInternalSegmentA
	SelectableInternalSegmentB
)

// Selectable represents something a user can pick in the editor. Initially this
// was implemented via just passing `any` instances around, but that had the
// limitation of not being able to select parts of objects that weren't explicit
// types. (for example, only one point of an internal segment)
type Selectable struct {
	Type            EditorSelectableType
	Sector          *core.Sector
	Body            *core.Body
	SectorSegment   *core.SectorSegment
	InternalSegment *core.InternalSegment
}

func SelectableFromSector(s *core.Sector) *Selectable {
	return &Selectable{Type: SelectableSector, Sector: s}
}

func SelectableFromSegment(s *core.SectorSegment) *Selectable {
	return &Selectable{Type: SelectableSectorSegment, Sector: s.Sector, SectorSegment: s}
}

func SelectableFromBody(b *core.Body) *Selectable {
	return &Selectable{Type: SelectableBody, Sector: b.Sector(), Body: b}
}

func SelectableFromInternalSegment(s *core.InternalSegment) *Selectable {
	return &Selectable{Type: SelectableInternalSegment, Sector: s.Sector(), InternalSegment: s}
}

func SelectableFromInternalSegmentA(s *core.InternalSegment) *Selectable {
	return &Selectable{Type: SelectableInternalSegmentA, Sector: s.Sector(), InternalSegment: s}
}

func SelectableFromInternalSegmentB(s *core.InternalSegment) *Selectable {
	return &Selectable{Type: SelectableInternalSegmentB, Sector: s.Sector(), InternalSegment: s}
}

func SelectableFromEntityRef(ref *concepts.EntityRef) *Selectable {
	if sector := core.SectorFromDb(ref); sector != nil {
		return SelectableFromSector(sector)
	}
	if body := core.BodyFromDb(ref); body != nil {
		return SelectableFromBody(body)
	}
	if seg := core.InternalSegmentFromDb(ref); seg != nil {
		return SelectableFromInternalSegment(seg)
	}
	return nil
}

func (target *Selectable) IndexOf(list []*Selectable) int {
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
		}
		return i
	}
	return -1
}
func (s *Selectable) AddToList(list *[]*Selectable) bool {
	if s.IndexOf(*list) >= 0 {
		return false
	}
	*list = append(*list, s)
	return false
}

// Serialize saves the data for whatever the selectable is holding, which may or
// may not be an Entity (could be a component of one)
func (s *Selectable) Serialize() any {
	switch s.Type {
	case SelectableSectorSegment:
		fallthrough
	case SelectableSector:
		return s.Sector.DB.SerializeEntity(s.Sector.Entity)
	case SelectableBody:
		return s.Body.DB.SerializeEntity(s.Body.Entity)
	case SelectableInternalSegment:
		fallthrough
	case SelectableInternalSegmentA:
		fallthrough
	case SelectableInternalSegmentB:
		return s.InternalSegment.DB.SerializeEntity(s.InternalSegment.Entity)
	}
	return nil
}
