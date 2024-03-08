package state

import (
	"math"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
)

type ColumnPortal struct {
	*Column
	Adj                 *core.Sector
	AdjSegment          *core.SectorSegment
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
	cp.Adj = core.SectorFromDb(cp.SectorSegment.AdjacentSector)
	cp.AdjSegment = cp.SectorSegment.AdjacentSegment
	cp.AdjFloorZ, cp.AdjCeilZ = cp.Adj.SlopedZRender(cp.RaySegIntersect.To2D())
	cp.AdjProjHeightTop = cp.ProjectZ(cp.AdjCeilZ - cp.CameraZ)
	cp.AdjProjHeightBottom = cp.ProjectZ(cp.AdjFloorZ - cp.CameraZ)
	cp.AdjScreenTop = cp.ScreenHeight/2 - int(math.Floor(cp.AdjProjHeightTop))
	cp.AdjScreenBottom = cp.ScreenHeight/2 - int(math.Floor(cp.AdjProjHeightBottom))
	cp.AdjClippedTop = concepts.Clamp(cp.AdjScreenTop, cp.ClippedStart, cp.ClippedEnd)
	cp.AdjClippedBottom = concepts.Clamp(cp.AdjScreenBottom, cp.ClippedStart, cp.ClippedEnd)
}
