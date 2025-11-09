// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package core

import (
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/dynamic"
	"tlyakhov/gofoom/ecs"

	"github.com/spf13/cast"
)

const (
	matchEpsilon float64 = 1e-4
)

type SectorSegment struct {
	Segment `editable:"^"`

	P         dynamic.DynamicValue[concepts.Vector2] `editable:"X/Y"`
	LoSurface materials.Surface                      `editable:"Low"`
	HiSurface materials.Surface                      `editable:"High"`
	// TODO: This should also be implemented for hi/lo surfaces
	WallUVIgnoreSlope bool `editable:"Wall U/V ignore slope"`
	PortalHasMaterial bool `editable:"Portal has material"`
	PortalIsPassable  bool `editable:"Portal is passable"`
	PortalTeleports   bool `editable:"Portal sector not adjacent"`

	AdjacentSector  ecs.Entity     `editable:"Portal sector" edit_type:"Sector"`
	AdjacentSegment *SectorSegment `ecs:"norelation"`
	// Only when loading or linking
	AdjacentSegmentIndex int `editable:"Portal segment index"`

	Sector *Sector

	// Pre-calculated attributes
	Index              int
	Next               *SectorSegment `ecs:"norelation"`
	Prev               *SectorSegment `ecs:"norelation"`
	PortalMatrix       concepts.Matrix2
	MirrorPortalMatrix concepts.Matrix2

	Flags int
}

func (s *SectorSegment) Recalculate() {
	s.Segment.A = &s.P.Render
	if s.Next != nil {
		s.Segment.B = &s.Next.P.Render
	}
	s.Segment.Recalculate()

	if s.Sector != nil && s.Sector.Winding < 0 {
		s.Segment.Normal.MulSelf(-1)
	}

	// These are for transforming coordinate spaces through teleporting portals
	// TODO: A waste to store these for non-portal segments
	s.PortalMatrix[concepts.MatBasis1X] = s.B[0] - s.A[0]
	s.PortalMatrix[concepts.MatBasis1Y] = s.B[1] - s.A[1]
	s.PortalMatrix[concepts.MatBasis2X] = s.Normal[0]
	s.PortalMatrix[concepts.MatBasis2Y] = s.Normal[1]
	s.PortalMatrix[concepts.MatTransX] = s.A[0]
	s.PortalMatrix[concepts.MatTransY] = s.A[1]
	s.MirrorPortalMatrix[concepts.MatBasis1X] = -(s.B[0] - s.A[0])
	s.MirrorPortalMatrix[concepts.MatBasis1Y] = -(s.B[1] - s.A[1])
	s.MirrorPortalMatrix[concepts.MatBasis2X] = -s.Normal[0]
	s.MirrorPortalMatrix[concepts.MatBasis2Y] = -s.Normal[1]
	s.MirrorPortalMatrix[concepts.MatTransX] = s.B[0]
	s.MirrorPortalMatrix[concepts.MatTransY] = s.B[1]

	if s.Sector != nil {
		s.RealizeAdjacentSector()
	}
}

func (s *SectorSegment) RealizeAdjacentSector() {
	if s.AdjacentSector == 0 {
		return
	}

	if adj := GetSector(s.AdjacentSector); adj != nil {
		// Get the actual segment using the index
		if s.PortalTeleports && s.AdjacentSegmentIndex != -1 {
			s.AdjacentSegment = adj.Segments[s.AdjacentSegmentIndex]
			return
		}
		// Get the actual segment by finding a matching one
		for _, s2 := range adj.Segments {
			if s2.Matches(&s.Segment) {
				s.AdjacentSegment = s2
				break
			}
		}
		// Pick one, to avoid nil reference errors
		if s.AdjacentSegment == nil {
			s.AdjacentSegment = adj.Segments[0]
		}
	}
}

func (s *SectorSegment) Split(p concepts.Vector2) *SectorSegment {
	// Segments are linked list where the line goes from `s.P` -> `s.Next.P`
	// The insertion index for the split should therefore be such that `s` is
	// the _previous_ segment
	index := 0
	for index < len(s.Sector.Segments) &&
		s.Sector.Segments[index].Prev != s {
		index++
	}
	// Make a copy to preserve everything other than the point.
	copied := new(SectorSegment)
	copied.Sector = s.Sector
	copied.Construct(s.Serialize())
	copied.P.SetAll(p)
	// Insert into the sector
	s.Sector.Segments = append(s.Sector.Segments[:index+1], s.Sector.Segments[index:]...)
	s.Sector.Segments[index] = copied
	// Recalculate metadata
	//DbgPrintSegments(seg.Sector)
	s.Sector.Recalculate()
	//DbgPrintSegments(seg.Sector)
	return copied
}

func (s *SectorSegment) Construct(data map[string]any) {
	s.Segment.Construct(data)
	s.P.Construct(nil)
	s.Normal = concepts.Vector2{}
	s.WallUVIgnoreSlope = false
	s.PortalHasMaterial = false
	s.PortalIsPassable = true
	s.PortalTeleports = false
	s.HiSurface.Construct(nil)
	s.LoSurface.Construct(nil)
	s.AdjacentSegmentIndex = -1

	if data == nil {
		return
	}

	if v, ok := data["P"]; ok {
		s.P.Construct(v)
	}
	if v, ok := data["WallUVIgnoreSlope"]; ok {
		s.WallUVIgnoreSlope = v.(bool)
	}
	if v, ok := data["PortalHasMaterial"]; ok {
		s.PortalHasMaterial = v.(bool)
	}
	if v, ok := data["PortalIsPassable"]; ok {
		s.PortalIsPassable = v.(bool)
	}
	if v, ok := data["PortalTeleports"]; ok {
		s.PortalTeleports = v.(bool)
	}
	if v, ok := data["AdjacentSector"]; ok {
		s.AdjacentSector, _ = ecs.ParseEntity(v.(string))
	}
	if v, ok := data["AdjacentSegment"]; ok {
		if parsed, err := cast.ToIntE(v); err == nil {
			s.AdjacentSegmentIndex = parsed
		}
	}
	if v, ok := data["Lo"]; ok {
		s.LoSurface.Construct(v.(map[string]any))
	}
	if v, ok := data["Hi"]; ok {
		s.HiSurface.Construct(v.(map[string]any))
	}
}

func (s *SectorSegment) Serialize() map[string]any {
	result := s.Segment.Serialize()
	// These are pointers to s.P and s.Next.P, don't save them.
	delete(result, "A")
	delete(result, "B")

	result["P"] = s.P.Serialize()

	if s.WallUVIgnoreSlope {
		result["WallUVIgnoreSlope"] = true
	}
	if s.PortalHasMaterial {
		result["PortalHasMaterial"] = true
	}
	if !s.PortalIsPassable {
		result["PortalIsPassable"] = false
	}
	if s.PortalTeleports {
		result["PortalTeleports"] = true
		if s.AdjacentSegment != nil {
			result["AdjacentSegment"] = s.AdjacentSegment.Index
		}
	}

	result["Lo"] = s.LoSurface.Serialize()
	result["Hi"] = s.HiSurface.Serialize()

	if s.AdjacentSector != 0 {
		result["AdjacentSector"] = s.AdjacentSector.Serialize()
	}

	return result
}
