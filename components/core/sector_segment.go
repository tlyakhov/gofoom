package core

import (
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"
)

const (
	matchEpsilon float64 = 1e-4
)

type SectorSegment struct {
	DB *concepts.EntityComponentDB
	Segment

	P                 concepts.Vector2  `editable:"X/Y"`
	LoSurface         materials.Surface `editable:"Low Surface"`
	HiSurface         materials.Surface `editable:"High Surface"`
	PortalHasMaterial bool              `editable:"Portal has material"`
	PortalIsPassable  bool              `editable:"Portal is passable"`

	AdjacentSector  *concepts.EntityRef
	AdjacentSegment *SectorSegment

	Sector *Sector
	Next   *SectorSegment
	Prev   *SectorSegment

	Flags int
}

func (s *SectorSegment) Recalculate() {
	if s.Sector != nil {
		s.Segment.Top = s.Sector.Max[2]
		s.Segment.Bottom = s.Sector.Min[2]
	}
	s.Segment.A = &s.P
	if s.Next != nil {
		s.Segment.B = &s.Next.P
	}
	s.Segment.Recalculate()
	if s.Sector != nil {
		s.RealizeAdjacentSector()
	}
}

func (s *SectorSegment) RealizeAdjacentSector() {
	if s.AdjacentSector.Nil() {
		return
	}

	if adj, ok := s.AdjacentSector.Component(SectorComponentIndex).(*Sector); ok {
		// Get the actual segment
		for _, s2 := range adj.Segments {
			if s2.Matches(&s.Segment) {
				s.AdjacentSegment = s2
				break
			}
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
	copied.DB = s.DB
	copied.Construct(s.Serialize())
	copied.P = p
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
	s.Segment.Construct(s.DB, data)
	s.P = concepts.Vector2{}
	s.Normal = concepts.Vector2{}
	s.PortalHasMaterial = false
	s.PortalIsPassable = true

	if data == nil {
		return
	}

	if v, ok := data["X"]; ok {
		s.P[0] = v.(float64)
	}
	if v, ok := data["Y"]; ok {
		s.P[1] = v.(float64)
	}
	if v, ok := data["PortalHasMaterial"]; ok {
		s.PortalHasMaterial = v.(bool)
	}
	if v, ok := data["PortalIsPassable"]; ok {
		s.PortalIsPassable = v.(bool)
	}
	if v, ok := data["AdjacentSector"]; ok {
		s.AdjacentSector = s.DB.DeserializeEntityRef(v)
	}
	if v, ok := data["Lo"]; ok {
		s.LoSurface.Construct(s.DB, v.(map[string]any))
	}
	if v, ok := data["Hi"]; ok {
		s.HiSurface.Construct(s.DB, v.(map[string]any))
	}
}

func (s *SectorSegment) Serialize() map[string]any {
	result := s.Segment.Serialize(false)
	result["X"] = s.P[0]
	result["Y"] = s.P[1]

	if s.PortalHasMaterial {
		result["PortalHasMaterial"] = true
	}
	if !s.PortalIsPassable {
		result["PortalIsPassable"] = false
	}

	result["Lo"] = s.LoSurface.Serialize()
	result["Hi"] = s.HiSurface.Serialize()

	if !s.AdjacentSector.Nil() {
		result["AdjacentSector"] = s.AdjacentSector.Serialize()
	}

	return result
}
