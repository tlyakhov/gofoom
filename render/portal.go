package render

import (
	"github.com/tlyakhov/gofoom/mapping"
	"github.com/tlyakhov/gofoom/render/material"
	"github.com/tlyakhov/gofoom/render/state"
)

func WallHi(slice *state.SlicePortal) {
	mat := material.For(slice.Segment.HiMaterial, slice.Slice)

	for slice.Y = slice.ClippedStart; slice.Y < slice.AdjClippedTop; slice.Y++ {
		screenIndex := uint(slice.X + slice.Y*slice.ScreenWidth)
		//fmt.Printf("%v %v %v %v\n", screenIndex, slice.Y, slice.ClippedStart, slice.AdjClippedTop)
		if slice.Distance >= slice.ZBuffer[screenIndex] {
			continue
		}
		v := float64(slice.Y-slice.ScreenStart) / float64(slice.AdjScreenTop-slice.ScreenStart)
		slice.Intersection.Z = slice.PhysicalSector.TopZ - v*(slice.PhysicalSector.TopZ-slice.Adj.GetPhysical().TopZ)

		// var light = this.map.light(slice.intersection, segment.normal, slice.sector, slice.segment, slice.u, v * 0.5, true);

		if slice.AdjSegment.HiBehavior == mapping.ScaleWidth || slice.AdjSegment.HiBehavior == mapping.ScaleNone {
			v = (v*(slice.Adj.GetPhysical().TopZ-slice.PhysicalSector.TopZ) - slice.Adj.GetPhysical().TopZ) / 64.0
		}
		if mat != nil {
			slice.Write(screenIndex, mat.Sample(slice.U, v, nil, slice.ProjectZ(1.0)))
		}
		slice.ZBuffer[screenIndex] = slice.Distance
	}
}

func WallLow(slice *state.SlicePortal) {
	mat := material.For(slice.Segment.LoMaterial, slice.Slice)

	for slice.Y = slice.AdjClippedBottom; slice.Y < slice.ClippedEnd; slice.Y++ {
		screenIndex := uint(slice.X + slice.Y*slice.ScreenWidth)
		if slice.Distance >= slice.ZBuffer[screenIndex] {
			continue
		}
		v := float64(slice.Y-slice.AdjClippedBottom) / float64(slice.ScreenEnd-slice.AdjScreenBottom)
		slice.Intersection.Z = slice.Adj.GetPhysical().BottomZ - v*(slice.Adj.GetPhysical().BottomZ-slice.PhysicalSector.BottomZ)
		// var light = this.map.light(slice.intersection, segment.normal, slice.sector, slice.segment, slice.u, v * 0.5 + 0.5, true);
		if slice.AdjSegment.LoBehavior == mapping.ScaleWidth || slice.AdjSegment.LoBehavior == mapping.ScaleNone {
			v = (v*(slice.PhysicalSector.BottomZ-slice.Adj.GetPhysical().BottomZ) - slice.PhysicalSector.BottomZ) / 64.0
		}

		if mat != nil {
			slice.Write(screenIndex, mat.Sample(slice.U, v, nil, slice.ProjectZ(1.0)))
		}
		slice.ZBuffer[screenIndex] = slice.Distance
	}
}
