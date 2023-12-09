package state

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
)

type ColumnPortal struct {
	*Column
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

func (cp *ColumnPortal) CalcScreen() {
	cp.Adj = core.SectorFromDb(cp.Segment.AdjacentSector)
	cp.AdjSegment = cp.Segment.AdjacentSegment
	cp.AdjFloorZ, cp.AdjCeilZ = cp.Adj.SlopedZRender(cp.Intersection.To2D())
	cp.AdjProjHeightTop = cp.ProjectZ(cp.AdjCeilZ - cp.CameraZ)
	cp.AdjProjHeightBottom = cp.ProjectZ(cp.AdjFloorZ - cp.CameraZ)
	cp.AdjScreenTop = cp.ScreenHeight/2 - int(cp.AdjProjHeightTop)
	cp.AdjScreenBottom = cp.ScreenHeight/2 - int(cp.AdjProjHeightBottom)
	cp.AdjClippedTop = concepts.IntClamp(cp.AdjScreenTop, cp.ClippedStart, cp.ClippedEnd)
	cp.AdjClippedBottom = concepts.IntClamp(cp.AdjScreenBottom, cp.ClippedStart, cp.ClippedEnd)
}
