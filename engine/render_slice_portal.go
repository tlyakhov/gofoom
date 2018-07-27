package engine

import "github.com/tlyakhov/gofoom/util"

type RenderSlicePortal struct {
	*RenderSlice
	Adj                 *MapSector
	AdjSegment          *MapSegment
	AdjProjHeightTop    float64
	AdjProjHeightBottom float64
	AdjScreenTop        int
	AdjScreenBottom     int
	AdjClippedTop       int
	AdjClippedBottom    int
}

func (slice *RenderSlicePortal) CalcScreen() {
	slice.Adj = slice.Segment.AdjacentSector
	slice.AdjSegment = slice.Segment.AdjacentSegment
	slice.AdjProjHeightTop = slice.ProjectZ(slice.Adj.TopZ - slice.CameraZ)
	slice.AdjProjHeightBottom = slice.ProjectZ(slice.Adj.BottomZ - slice.CameraZ)
	slice.AdjScreenTop = slice.ScreenHeight/2 - int(slice.AdjProjHeightTop)
	slice.AdjScreenBottom = slice.ScreenHeight/2 - int(slice.AdjProjHeightBottom)
	slice.AdjClippedTop = util.Max(slice.AdjScreenTop, slice.ClippedStart)
	slice.AdjClippedBottom = util.Min(slice.AdjScreenBottom, slice.ClippedEnd)
}

func (slice *RenderSlicePortal) RenderHigh() {
	if slice.AdjSegment.HiMaterial == nil {
		return
	}

	for slice.Y = slice.ClippedStart; slice.Y < slice.AdjClippedTop; slice.Y++ {
		screenIndex := uint(slice.TargetX + slice.Y*slice.WorkerWidth)
		if slice.Distance >= slice.zbuffer[screenIndex] {
			continue
		}
		v := float64(slice.Y-slice.ScreenStart) / float64(slice.AdjScreenTop-slice.ScreenStart)
		slice.Intersection.Z = slice.Sector.TopZ - v*(slice.Sector.TopZ-slice.Adj.TopZ)

		// var light = this.map.light(slice.intersection, segment.normal, slice.sector, slice.segment, slice.u, v * 0.5, true);

		if slice.AdjSegment.HiBehavior == ScaleWidth || slice.AdjSegment.HiBehavior == ScaleNone {
			v = (v*(slice.Adj.TopZ-slice.Sector.TopZ) - slice.Adj.TopZ) / 64.0
		}
		slice.Write(screenIndex, slice.Segment.HiMaterial.Sample(slice.RenderSlice, slice.U, v, nil, uint(slice.AdjScreenTop-slice.ScreenStart)))
		slice.zbuffer[screenIndex] = slice.Distance
	}

}

func (slice *RenderSlicePortal) RenderLow() {
	if slice.AdjSegment.LoMaterial == nil {
		return
	}
	for slice.Y = slice.AdjClippedBottom; slice.Y < slice.ClippedEnd; slice.Y++ {
		screenIndex := uint(slice.TargetX + slice.Y*slice.WorkerWidth)
		if slice.Distance >= slice.zbuffer[screenIndex] {
			continue
		}
		v := float64(slice.Y-slice.AdjClippedBottom) / float64(slice.ScreenEnd-slice.AdjScreenBottom)
		slice.Intersection.Z = slice.Adj.BottomZ - v*(slice.Adj.BottomZ-slice.Sector.BottomZ)
		// var light = this.map.light(slice.intersection, segment.normal, slice.sector, slice.segment, slice.u, v * 0.5 + 0.5, true);
		if slice.AdjSegment.LoBehavior == ScaleWidth || slice.AdjSegment.LoBehavior == ScaleNone {
			v = (v*(slice.Sector.BottomZ-slice.Adj.BottomZ) - slice.Sector.BottomZ) / 64.0
		}

		slice.Write(screenIndex, slice.Segment.LoMaterial.Sample(slice.RenderSlice, slice.U, v, nil, uint(slice.ScreenEnd-slice.AdjScreenBottom)))
		slice.zbuffer[screenIndex] = slice.Distance
	}
}
