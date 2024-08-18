// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package core

import (
	"math"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/ecs"
)

//go:generate go run github.com/dmarkham/enumer -type=SelectableType -json
type SelectableType int

const (
	SelectableEntity SelectableType = iota
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
	SelectableBody:   SelectableBody,
	SelectableEntity: SelectableEntity,
}

// Selectable represents something an editor or player can pick. Initially this
// was implemented via just passing `any` instances around, but that had the
// limitation of not being able to select parts of objects that weren't explicit
// types. (for example, only one point of an internal segment)
type Selectable struct {
	ecs.Entity
	DB              *ecs.ECS
	Type            SelectableType
	Sector          *Sector
	Body            *Body
	SectorSegment   *SectorSegment
	InternalSegment *InternalSegment
}

func (s *Selectable) Hash() uint64 {
	if s.SectorSegment != nil {
		// 4 bits for type, 16 bits for segment index, 44 bits for entity
		return (uint64(s.Type) << 60) | uint64(s.SectorSegment.Index<<44) | uint64(s.Entity)
	}
	// 4 bits for type, 60 bits for entity
	return (uint64(s.Type) << 60) | uint64(s.Entity)
}

func (s *Selectable) GroupHash() uint64 {
	if s.SectorSegment != nil {
		// 4 bits for type, 16 bits for segment index, 44 bits for entity
		return (uint64(SelectableSectorSegment) << 60) | uint64(s.SectorSegment.Index<<44) | uint64(s.Entity)
	}
	// 4 bits for type, 60 bits for entity
	return (uint64(typeGroups[s.Type]) << 60) | uint64(s.Entity)
}

func SelectableFromSector(s *Sector) *Selectable {
	return &Selectable{Type: SelectableSector, Sector: s, Entity: s.Entity, DB: s.DB}
}

func SelectableFromFloor(s *Sector) *Selectable {
	return &Selectable{Type: SelectableFloor, Sector: s, Entity: s.Entity, DB: s.DB}
}

func SelectableFromCeil(s *Sector) *Selectable {
	return &Selectable{Type: SelectableCeiling, Sector: s, Entity: s.Entity, DB: s.DB}
}

func SelectableFromSegment(s *SectorSegment) *Selectable {
	return &Selectable{
		Type:          SelectableSectorSegment,
		Sector:        s.Sector,
		SectorSegment: s,
		Entity:        s.Sector.Entity,
		DB:            s.DB}
}

func SelectableFromWall(s *SectorSegment, t SelectableType) *Selectable {
	return &Selectable{
		Type:          t,
		Sector:        s.Sector,
		SectorSegment: s,
		Entity:        s.Sector.Entity,
		DB:            s.DB}
}

func SelectableFromBody(b *Body) *Selectable {
	return &Selectable{
		Type:   SelectableBody,
		Sector: b.Sector(),
		Body:   b,
		Entity: b.Entity,
		DB:     b.DB}
}

func SelectableFromInternalSegment(s *InternalSegment) *Selectable {
	return &Selectable{
		Type:            SelectableInternalSegment,
		InternalSegment: s,
		Entity:          s.Entity,
		DB:              s.DB}
}

func SelectableFromInternalSegmentA(s *InternalSegment) *Selectable {
	return &Selectable{Type: SelectableInternalSegmentA,
		InternalSegment: s,
		Entity:          s.Entity,
		DB:              s.DB}
}

func SelectableFromInternalSegmentB(s *InternalSegment) *Selectable {
	return &Selectable{Type: SelectableInternalSegmentB,
		InternalSegment: s,
		Entity:          s.Entity,
		DB:              s.DB}
}

func SelectableFromEntity(db *ecs.ECS, e ecs.Entity) *Selectable {
	if sector := SectorFromDb(db, e); sector != nil {
		return SelectableFromSector(sector)
	}
	if body := BodyFromDb(db, e); body != nil {
		return SelectableFromBody(body)
	}
	if seg := InternalSegmentFromDb(db, e); seg != nil {
		return SelectableFromInternalSegment(seg)
	}
	return &Selectable{Type: SelectableEntity, Entity: e, DB: db}
}

// Serialize saves the data for whatever the selectable is holding, which may or
// may not be an Entity (could be a component of one)
func (s *Selectable) Serialize() any {
	return s.DB.SerializeEntity(s.Entity)
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
		s.Body.Pos.ResetToOriginal()
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
		s.Body.Pos.ResetToOriginal()
	case SelectableInternalSegmentA:
		fallthrough
	case SelectableInternalSegmentB:
		fallthrough
	case SelectableInternalSegment:
		s.InternalSegment.Recalculate()
	}
}

func (s *Selectable) PointZ(p *concepts.Vector2) (bottom float64, top float64) {
	switch s.Type {
	case SelectableHi:
		_, top = s.Sector.PointZ(ecs.DynamicNow, p)
		adj := s.SectorSegment.AdjacentSegment.Sector
		_, adjTop := adj.PointZ(ecs.DynamicNow, p)
		bottom, top = math.Min(adjTop, top), math.Max(adjTop, top)
	case SelectableLow:
		bottom, _ = s.Sector.PointZ(ecs.DynamicNow, p)
		adj := s.SectorSegment.AdjacentSegment.Sector
		adjBottom, _ := adj.PointZ(ecs.DynamicNow, p)
		bottom, top = math.Min(adjBottom, bottom), math.Max(adjBottom, bottom)
	case SelectableMid:
		bottom, top = s.Sector.PointZ(ecs.DynamicNow, p)
	case SelectableInternalSegment:
		bottom, top = s.InternalSegment.Bottom, s.InternalSegment.Top
	}
	return
}
