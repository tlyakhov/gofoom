package state

import (
	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/mapping"
)

type SlicePortal struct {
	*Slice
	Adj                 mapping.AbstractSector
	AdjSegment          *mapping.Segment
	AdjProjHeightTop    float64
	AdjProjHeightBottom float64
	AdjScreenTop        int
	AdjScreenBottom     int
	AdjClippedTop       int
	AdjClippedBottom    int
}

func (slice *SlicePortal) CalcScreen() {
	slice.Adj = slice.Segment.AdjacentSector
	slice.AdjSegment = slice.Segment.AdjacentSegment
	slice.AdjProjHeightTop = slice.ProjectZ(slice.Adj.GetPhysical().TopZ - slice.CameraZ)
	slice.AdjProjHeightBottom = slice.ProjectZ(slice.Adj.GetPhysical().BottomZ - slice.CameraZ)
	slice.AdjScreenTop = slice.ScreenHeight/2 - int(slice.AdjProjHeightTop)
	slice.AdjScreenBottom = slice.ScreenHeight/2 - int(slice.AdjProjHeightBottom)
	slice.AdjClippedTop = concepts.IntClamp(slice.AdjScreenTop, slice.ClippedStart, slice.ClippedEnd)
	slice.AdjClippedBottom = concepts.IntClamp(slice.AdjScreenBottom, slice.ClippedStart, slice.ClippedEnd)
}
