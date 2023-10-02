package state

import (
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/core"
)

type SlicePortal struct {
	*Slice
	Adj                 core.AbstractSector
	AdjSegment          *core.Segment
	AdjProjHeightTop    float64
	AdjProjHeightBottom float64
	AdjFloorZ           float64
	AdjCeilZ            float64
	AdjScreenTop        int
	AdjScreenBottom     int
	AdjClippedTop       int
	AdjClippedBottom    int
}

func (s *SlicePortal) CalcScreen() {
	s.Adj = s.Segment.AdjacentSector
	s.AdjSegment = s.Segment.AdjacentSegment
	s.AdjFloorZ, s.AdjCeilZ = s.Adj.Physical().CalcFloorCeilingZ(s.Intersection.To2D())
	s.AdjProjHeightTop = s.ProjectZ(s.AdjCeilZ - s.CameraZ)
	s.AdjProjHeightBottom = s.ProjectZ(s.AdjFloorZ - s.CameraZ)
	s.AdjScreenTop = s.ScreenHeight/2 - int(s.AdjProjHeightTop)
	s.AdjScreenBottom = s.ScreenHeight/2 - int(s.AdjProjHeightBottom)
	s.AdjClippedTop = concepts.IntClamp(s.AdjScreenTop, s.ClippedStart, s.ClippedEnd)
	s.AdjClippedBottom = concepts.IntClamp(s.AdjScreenBottom, s.ClippedStart, s.ClippedEnd)
}
