package state

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
)

type SlicePortal struct {
	*Slice
	Adj                 *core.Sector
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
	s.Adj = core.SectorFromDb(s.Segment.AdjacentSector)
	s.AdjSegment = s.Segment.AdjacentSegment
	s.AdjFloorZ, s.AdjCeilZ = s.Adj.SlopedZRender(s.Intersection.To2D())
	s.AdjProjHeightTop = s.ProjectZ(s.AdjCeilZ - s.CameraZ)
	s.AdjProjHeightBottom = s.ProjectZ(s.AdjFloorZ - s.CameraZ)
	s.AdjScreenTop = s.ScreenHeight/2 - int(s.AdjProjHeightTop)
	s.AdjScreenBottom = s.ScreenHeight/2 - int(s.AdjProjHeightBottom)
	s.AdjClippedTop = concepts.IntClamp(s.AdjScreenTop, s.ClippedStart, s.ClippedEnd)
	s.AdjClippedBottom = concepts.IntClamp(s.AdjScreenBottom, s.ClippedStart, s.ClippedEnd)
}
