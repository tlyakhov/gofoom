package render

import (
	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/mapping"
)

type SlicePortal struct {
	*Slice
	Adj                 *mapping.Sector
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
	slice.AdjProjHeightTop = slice.ProjectZ(slice.Adj.TopZ - slice.CameraZ)
	slice.AdjProjHeightBottom = slice.ProjectZ(slice.Adj.BottomZ - slice.CameraZ)
	slice.AdjScreenTop = slice.ScreenHeight/2 - int(slice.AdjProjHeightTop)
	slice.AdjScreenBottom = slice.ScreenHeight/2 - int(slice.AdjProjHeightBottom)
	slice.AdjClippedTop = concepts.Max(slice.AdjScreenTop, slice.ClippedStart)
	slice.AdjClippedBottom = concepts.Min(slice.AdjScreenBottom, slice.ClippedEnd)
}

func (slice *SlicePortal) RenderHigh() {
	if slice.AdjSegment.HiMaterial == nil {
		return
	}

	for slice.Y = slice.ClippedStart; slice.Y < slice.AdjClippedTop; slice.Y++ {
		screenIndex := uint(slice.TargetX + slice.Y*slice.WorkerWidth)
		if slice.Distance >= slice.ZBuffer[screenIndex] {
			continue
		}
		v := float64(slice.Y-slice.ScreenStart) / float64(slice.AdjScreenTop-slice.ScreenStart)
		slice.Intersection.Z = slice.Sector.TopZ - v*(slice.Sector.TopZ-slice.Adj.TopZ)

		// var light = this.map.light(slice.intersection, segment.normal, slice.sector, slice.segment, slice.u, v * 0.5, true);

		if slice.AdjSegment.HiBehavior == mapping.ScaleWidth || slice.AdjSegment.HiBehavior == mapping.ScaleNone {
			v = (v*(slice.Adj.TopZ-slice.Sector.TopZ) - slice.Adj.TopZ) / 64.0
		}
		mat := concepts.Local(slice.Segment.HiMaterial, typeMap).(ISampler)
		slice.Write(screenIndex, mat.Sample(slice.Slice, slice.U, v, nil, slice.ProjectZ(1.0)))
		slice.ZBuffer[screenIndex] = slice.Distance
	}
}

func (slice *SlicePortal) RenderLow() {
	if slice.AdjSegment.LoMaterial == nil {
		return
	}
	for slice.Y = slice.AdjClippedBottom; slice.Y < slice.ClippedEnd; slice.Y++ {
		screenIndex := uint(slice.TargetX + slice.Y*slice.WorkerWidth)
		if slice.Distance >= slice.ZBuffer[screenIndex] {
			continue
		}
		v := float64(slice.Y-slice.AdjClippedBottom) / float64(slice.ScreenEnd-slice.AdjScreenBottom)
		slice.Intersection.Z = slice.Adj.BottomZ - v*(slice.Adj.BottomZ-slice.Sector.BottomZ)
		// var light = this.map.light(slice.intersection, segment.normal, slice.sector, slice.segment, slice.u, v * 0.5 + 0.5, true);
		if slice.AdjSegment.LoBehavior == mapping.ScaleWidth || slice.AdjSegment.LoBehavior == mapping.ScaleNone {
			v = (v*(slice.Sector.BottomZ-slice.Adj.BottomZ) - slice.Sector.BottomZ) / 64.0
		}

		mat := concepts.Local(slice.Segment.LoMaterial, typeMap).(ISampler)
		slice.Write(screenIndex, mat.Sample(slice.Slice, slice.U, v, nil, slice.ProjectZ(1.0)))
		slice.ZBuffer[screenIndex] = slice.Distance
	}
}
