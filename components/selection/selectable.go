// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package selection

import (
	"math"
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
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
	return &Selectable{Type: SelectableSector, Sector: s, Entity: s.Entity}
}

func SelectableFromFloor(s *core.Sector) *Selectable {
	return &Selectable{Type: SelectableFloor, Sector: s, Entity: s.Entity}
}

func SelectableFromCeil(s *core.Sector) *Selectable {
	return &Selectable{Type: SelectableCeiling, Sector: s, Entity: s.Entity}
}

func SelectableFromSegment(s *core.SectorSegment) *Selectable {
	return &Selectable{
		Type:          SelectableSectorSegment,
		Sector:        s.Sector,
		SectorSegment: s,
		Entity:        s.Sector.Entity}
}

func SelectableFromWall(s *core.SectorSegment, t SelectableType) *Selectable {
	return &Selectable{
		Type:          t,
		Sector:        s.Sector,
		SectorSegment: s,
		Entity:        s.Sector.Entity}
}

func SelectableFromBody(b *core.Body) *Selectable {
	return &Selectable{
		Type:   SelectableBody,
		Sector: b.Sector(),
		Body:   b,
		Entity: b.Entity}
}

func SelectableFromInternalSegment(s *core.InternalSegment) *Selectable {
	return &Selectable{
		Type:            SelectableInternalSegment,
		InternalSegment: s,
		Entity:          s.Entity}
}

func SelectableFromInternalSegmentA(s *core.InternalSegment) *Selectable {
	return &Selectable{Type: SelectableInternalSegmentA,
		InternalSegment: s,
		Entity:          s.Entity}
}

func SelectableFromInternalSegmentB(s *core.InternalSegment) *Selectable {
	return &Selectable{Type: SelectableInternalSegmentB,
		InternalSegment: s,
		Entity:          s.Entity}
}

func SelectableFromActionWaypoint(s *behaviors.ActionWaypoint) *Selectable {
	return &Selectable{
		Type:           SelectableActionWaypoint,
		ActionWaypoint: s,
		Entity:         s.Entity}
}

func SelectableFromEntity(e ecs.Entity) *Selectable {
	if sector := core.GetSector(e); sector != nil {
		return SelectableFromSector(sector)
	}
	if body := core.GetBody(e); body != nil {
		return SelectableFromBody(body)
	}
	if seg := core.GetInternalSegment(e); seg != nil {
		return SelectableFromInternalSegment(seg)
	}
	if aw := behaviors.GetActionWaypoint(e); aw != nil {
		return SelectableFromActionWaypoint(aw)
	}
	return &Selectable{Type: SelectableEntity, Entity: e}
}

// Helpful for prserving selections after loading snapshots
func (s *Selectable) Refresh() bool {
	if s.ActionWaypoint != nil {
		s.ActionWaypoint = behaviors.GetActionWaypoint(s.Entity)
		if s.ActionWaypoint == nil {
			return false
		}
	}
	if s.Body != nil {
		s.Body = core.GetBody(s.Entity)
		if s.Body == nil {
			return false
		}
	}
	if s.InternalSegment != nil {
		s.InternalSegment = core.GetInternalSegment(s.Entity)
		if s.InternalSegment == nil {
			return false
		}
	}
	if s.Sector != nil {
		s.Sector = core.GetSector(s.Entity)
		if s.Sector == nil &&
			(s.Type == SelectableSector || s.Type == SelectableCeiling || s.Type == SelectableFloor) {
			return false
		}
	}
	if s.SectorSegment != nil {
		index := s.SectorSegment.Index
		if s.Sector == nil || index >= len(s.Sector.Segments) {
			return false
		}
		s.SectorSegment = s.Sector.Segments[index]
	}
	return true
}

// Serialize saves the data for whatever the selectable is holding, which may or
// may not be an Entity (could be a component of one)
func (s *Selectable) Serialize() any {
	return ecs.SerializeEntity(s.Entity, false)
}

func (s *Selectable) PositionRange(f func(p *concepts.Vector2)) {
	switch s.Type {
	case SelectableSector:
		for _, seg := range s.Sector.Segments {
			f(&seg.P.Spawn)
			seg.P.ResetToSpawn()
		}
		s.Sector.Precompute()
	case SelectableLow, SelectableMid, SelectableHi:
		fallthrough
	case SelectableSectorSegment:
		f(&s.SectorSegment.P.Spawn)
		s.SectorSegment.P.ResetToSpawn()
		s.Sector.Precompute()
	case SelectableActionWaypoint:
		f(s.ActionWaypoint.P.To2D())
	case SelectableBody:
		f(s.Body.Pos.Spawn.To2D())
		s.Body.Pos.ResetToSpawn()
	case SelectableInternalSegmentA:
		f(s.InternalSegment.A)
		s.InternalSegment.Precompute()
	case SelectableInternalSegmentB:
		f(s.InternalSegment.B)
		s.InternalSegment.Precompute()
	case SelectableInternalSegment:
		f(s.InternalSegment.A)
		f(s.InternalSegment.B)
		s.InternalSegment.Precompute()
	}
}

func (s *Selectable) Precompute() {
	switch s.Type {
	case SelectableSector, SelectableLow, SelectableMid, SelectableHi:
		fallthrough
	case SelectableSectorSegment:
		s.Sector.Precompute()
	case SelectablePath:
		fallthrough
	case SelectableBody:
		s.Body.Pos.ResetToSpawn()
	case SelectableInternalSegmentA, SelectableInternalSegmentB:
		fallthrough
	case SelectableInternalSegment:
		s.InternalSegment.Precompute()
	}
}

func (s *Selectable) PointZ(p *concepts.Vector2) (bottom float64, top float64) {
	switch s.Type {
	case SelectableHi:
		_, top = s.Sector.ZAt(p)
		adj := s.SectorSegment.AdjacentSegment.Sector
		_, adjTop := adj.ZAt(p)
		bottom, top = math.Min(adjTop, top), math.Max(adjTop, top)
	case SelectableLow:
		bottom, _ = s.Sector.ZAt(p)
		adj := s.SectorSegment.AdjacentSegment.Sector
		adjBottom, _ := adj.ZAt(p)
		bottom, top = math.Min(adjBottom, bottom), math.Max(adjBottom, bottom)
	case SelectableMid:
		bottom, top = s.Sector.ZAt(p)
	case SelectableInternalSegment:
		bottom, top = s.InternalSegment.Bottom, s.InternalSegment.Top
	}
	return
}
