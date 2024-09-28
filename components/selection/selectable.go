// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package selection

import (
	"math"
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/dynamic"
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
	SelectablePath
	SelectableActionWaypoint
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
	// Actions
	SelectableActionWaypoint: SelectableActionWaypoint,
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
	ECS             *ecs.ECS
	Type            SelectableType
	Sector          *core.Sector
	Body            *core.Body
	SectorSegment   *core.SectorSegment
	InternalSegment *core.InternalSegment
	ActionWaypoint  *behaviors.ActionWaypoint
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
		return s.Hash()
	}
	// 4 bits for type, 60 bits for entity
	return (uint64(typeGroups[s.Type]) << 60) | uint64(s.Entity)
}

func SelectableFromSector(s *core.Sector) *Selectable {
	return &Selectable{Type: SelectableSector, Sector: s, Entity: s.Entity, ECS: s.ECS}
}

func SelectableFromFloor(s *core.Sector) *Selectable {
	return &Selectable{Type: SelectableFloor, Sector: s, Entity: s.Entity, ECS: s.ECS}
}

func SelectableFromCeil(s *core.Sector) *Selectable {
	return &Selectable{Type: SelectableCeiling, Sector: s, Entity: s.Entity, ECS: s.ECS}
}

func SelectableFromSegment(s *core.SectorSegment) *Selectable {
	return &Selectable{
		Type:          SelectableSectorSegment,
		Sector:        s.Sector,
		SectorSegment: s,
		Entity:        s.Sector.Entity,
		ECS:           s.ECS}
}

func SelectableFromWall(s *core.SectorSegment, t SelectableType) *Selectable {
	return &Selectable{
		Type:          t,
		Sector:        s.Sector,
		SectorSegment: s,
		Entity:        s.Sector.Entity,
		ECS:           s.ECS}
}

func SelectableFromBody(b *core.Body) *Selectable {
	return &Selectable{
		Type:   SelectableBody,
		Sector: b.Sector(),
		Body:   b,
		Entity: b.Entity,
		ECS:    b.ECS}
}

func SelectableFromInternalSegment(s *core.InternalSegment) *Selectable {
	return &Selectable{
		Type:            SelectableInternalSegment,
		InternalSegment: s,
		Entity:          s.Entity,
		ECS:             s.ECS}
}

func SelectableFromInternalSegmentA(s *core.InternalSegment) *Selectable {
	return &Selectable{Type: SelectableInternalSegmentA,
		InternalSegment: s,
		Entity:          s.Entity,
		ECS:             s.ECS}
}

func SelectableFromInternalSegmentB(s *core.InternalSegment) *Selectable {
	return &Selectable{Type: SelectableInternalSegmentB,
		InternalSegment: s,
		Entity:          s.Entity,
		ECS:             s.ECS}
}

func SelectableFromActionWaypoint(s *behaviors.ActionWaypoint) *Selectable {
	return &Selectable{
		Type:           SelectableActionWaypoint,
		ActionWaypoint: s,
		Entity:         s.Entity,
		ECS:            s.ECS}
}

func SelectableFromEntity(db *ecs.ECS, e ecs.Entity) *Selectable {
	if sector := core.GetSector(db, e); sector != nil {
		return SelectableFromSector(sector)
	}
	if body := core.GetBody(db, e); body != nil {
		return SelectableFromBody(body)
	}
	if seg := core.GetInternalSegment(db, e); seg != nil {
		return SelectableFromInternalSegment(seg)
	}
	if aw := behaviors.GetActionWaypoint(db, e); aw != nil {
		return SelectableFromActionWaypoint(aw)
	}
	return &Selectable{Type: SelectableEntity, Entity: e, ECS: db}
}

// Serialize saves the data for whatever the selectable is holding, which may or
// may not be an Entity (could be a component of one)
func (s *Selectable) Serialize() any {
	return s.ECS.SerializeEntity(s.Entity)
}

func (s *Selectable) PositionRange(f func(p *concepts.Vector2)) {
	switch s.Type {
	case SelectableSector:
		for _, seg := range s.Sector.Segments {
			f(&seg.P)
		}
		s.Sector.Recalculate()
	case SelectableLow, SelectableMid, SelectableHi:
		fallthrough
	case SelectableSectorSegment:
		f(&s.SectorSegment.P)
		s.Sector.Recalculate()
	case SelectableActionWaypoint:
		f(s.ActionWaypoint.P.To2D())
	case SelectableBody:
		f(s.Body.Pos.Original.To2D())
		s.Body.Pos.ResetToOriginal()
	case SelectableInternalSegmentA:
		f(s.InternalSegment.A)
		s.InternalSegment.Recalculate()
	case SelectableInternalSegmentB:
		f(s.InternalSegment.B)
		s.InternalSegment.Recalculate()
	case SelectableInternalSegment:
		f(s.InternalSegment.A)
		f(s.InternalSegment.B)
		s.InternalSegment.Recalculate()
	}
}

func (s *Selectable) Recalculate() {
	switch s.Type {
	case SelectableSector, SelectableLow, SelectableMid, SelectableHi:
		fallthrough
	case SelectableSectorSegment:
		s.Sector.Recalculate()
	case SelectablePath:
		fallthrough
	case SelectableBody:
		s.Body.Pos.ResetToOriginal()
	case SelectableInternalSegmentA, SelectableInternalSegmentB:
		fallthrough
	case SelectableInternalSegment:
		s.InternalSegment.Recalculate()
	}
}

func (s *Selectable) PointZ(p *concepts.Vector2) (bottom float64, top float64) {
	switch s.Type {
	case SelectableHi:
		_, top = s.Sector.ZAt(dynamic.DynamicNow, p)
		adj := s.SectorSegment.AdjacentSegment.Sector
		_, adjTop := adj.ZAt(dynamic.DynamicNow, p)
		bottom, top = math.Min(adjTop, top), math.Max(adjTop, top)
	case SelectableLow:
		bottom, _ = s.Sector.ZAt(dynamic.DynamicNow, p)
		adj := s.SectorSegment.AdjacentSegment.Sector
		adjBottom, _ := adj.ZAt(dynamic.DynamicNow, p)
		bottom, top = math.Min(adjBottom, bottom), math.Max(adjBottom, bottom)
	case SelectableMid:
		bottom, top = s.Sector.ZAt(dynamic.DynamicNow, p)
	case SelectableInternalSegment:
		bottom, top = s.InternalSegment.Bottom, s.InternalSegment.Top
	}
	return
}
