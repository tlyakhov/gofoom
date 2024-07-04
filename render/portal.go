// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package render

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/render/state"
)

func WallHiPick(cp *state.ColumnPortal) {
	if cp.ScreenY >= cp.ClippedStart && cp.ScreenY < cp.AdjClippedTop {
		cp.PickedSelection = append(cp.PickedSelection, core.SelectableFromWall(cp.AdjSegment, core.SelectableHi))
	}
}

// WallHi renders the top portion of a portal segment.
func WallHi(cp *state.ColumnPortal) {
	surf := cp.AdjSegment.HiSurface
	u := cp.U
	if surf.Stretch == materials.StretchNone {
		if cp.Sector.Winding < 0 {
			u = 1.0 - u
		}
		u = (cp.Segment.A[0] + cp.Segment.A[1] + u*cp.Segment.Length) / 64.0
	}
	for cp.ScreenY = cp.ClippedStart; cp.ScreenY < cp.AdjClippedTop; cp.ScreenY++ {
		screenIndex := uint32(cp.ScreenX + cp.ScreenY*cp.ScreenWidth)
		if cp.Distance >= cp.ZBuffer[screenIndex] {
			continue
		}
		v := float64(cp.ScreenY-cp.ScreenStart) / float64(cp.AdjScreenTop-cp.ScreenStart)
		cp.RaySegIntersect[2] = (1.0-v)*cp.TopZ + v*cp.AdjCeilZ

		if surf.Stretch == materials.StretchNone {
			v = -cp.RaySegIntersect[2] / 64.0
		} else if surf.Stretch == materials.StretchAspect {
			v *= (cp.Sector.Max[2] - cp.Sector.Min[2]) / cp.Segment.Length
		}

		if surf.Material != 0 {
			tu := surf.Transform[0]*u + surf.Transform[2]*v + surf.Transform[4]
			tv := surf.Transform[1]*u + surf.Transform[3]*v + surf.Transform[5]
			cp.SampleShader(surf.Material, surf.ExtraStages, tu, tv, cp.ProjectZ(1.0))
			cp.SampleLight(&cp.MaterialSampler.Output, surf.Material, &cp.RaySegIntersect, cp.Distance)
		}
		//concepts.AsmVector4AddPreMulColorSelf((*[4]float64)(&s.FrameBuffer[screenIndex]), (*[4]float64)(&s.Material))
		cp.FrameBuffer[screenIndex].AddPreMulColorSelf(&cp.MaterialSampler.Output)
		cp.ZBuffer[screenIndex] = cp.Distance
	}
}

func WallLowPick(cp *state.ColumnPortal) {
	if cp.ScreenY >= cp.AdjClippedBottom && cp.ScreenY < cp.ClippedEnd {
		cp.PickedSelection = append(cp.PickedSelection, core.SelectableFromWall(cp.AdjSegment, core.SelectableLow))
	}
}

// WallLow renders the bottom portion of a portal segment.
func WallLow(cp *state.ColumnPortal) {
	surf := cp.AdjSegment.LoSurface
	u := cp.U
	if surf.Stretch == materials.StretchNone {
		if cp.Sector.Winding < 0 {
			u = 1.0 - u
		}
		u = (cp.Segment.A[0] + cp.Segment.A[1] + u*cp.Segment.Length) / 64.0
	}
	for cp.ScreenY = cp.AdjClippedBottom; cp.ScreenY < cp.ClippedEnd; cp.ScreenY++ {
		screenIndex := uint32(cp.ScreenX + cp.ScreenY*cp.ScreenWidth)
		if cp.Distance >= cp.ZBuffer[screenIndex] {
			continue
		}
		v := float64(cp.ScreenY-cp.AdjScreenBottom) / float64(cp.ScreenEnd-cp.AdjScreenBottom)
		cp.RaySegIntersect[2] = (1.0-v)*cp.AdjFloorZ + v*cp.BottomZ

		if surf.Stretch == materials.StretchNone {
			v = -cp.RaySegIntersect[2] / 64.0
		} else if surf.Stretch == materials.StretchAspect {
			v *= (cp.Sector.Max[2] - cp.Sector.Min[2]) / cp.Segment.Length
		}

		if surf.Material != 0 {
			tu := surf.Transform[0]*u + surf.Transform[2]*v + surf.Transform[4]
			tv := surf.Transform[1]*u + surf.Transform[3]*v + surf.Transform[5]
			cp.SampleShader(surf.Material, surf.ExtraStages, tu, tv, cp.ProjectZ(1.0))
			cp.SampleLight(&cp.MaterialSampler.Output, surf.Material, &cp.RaySegIntersect, cp.Distance)
		}
		//concepts.AsmVector4AddPreMulColorSelf((*[4]float64)(&s.FrameBuffer[screenIndex]), (*[4]float64)(&s.Material))
		cp.FrameBuffer[screenIndex].AddPreMulColorSelf(&cp.MaterialSampler.Output)
		cp.ZBuffer[screenIndex] = cp.Distance
	}
}
