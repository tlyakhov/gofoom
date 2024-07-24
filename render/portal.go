// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package render

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/render/state"
)

func WallHiPick(cp *state.ColumnPortal) {
	if cp.ScreenY >= cp.ClippedTop && cp.ScreenY < cp.AdjClippedTop {
		cp.PickedSelection = append(cp.PickedSelection, core.SelectableFromWall(cp.AdjSegment, core.SelectableHi))
	}
}

// WallHi renders the top portion of a portal segment.
func WallHi(cp *state.ColumnPortal) {
	surf := cp.AdjSegment.HiSurface
	transform := surf.Transform.Render
	// To calculate the vertical texture coordinate, we can't use the integer
	// screen coordinates, we need to use the precise floats
	vStart := float64(cp.ScreenHeight/2) - cp.ProjectedTop
	for cp.ScreenY = cp.ClippedTop; cp.ScreenY < cp.AdjClippedTop; cp.ScreenY++ {
		screenIndex := uint32(cp.ScreenX + cp.ScreenY*cp.ScreenWidth)
		if cp.Distance >= cp.ZBuffer[screenIndex] {
			continue
		}
		v := (float64(cp.ScreenY) - vStart) / (cp.ProjectedTop - cp.AdjProjectedTop)
		cp.RaySegIntersect[2] = (1.0-v)*cp.IntersectionTop + v*cp.AdjTop

		if surf.Material != 0 {
			tu := transform[0]*cp.U + transform[2]*v + transform[4]
			tv := transform[1]*cp.U + transform[3]*v + transform[5]
			cp.SampleShader(surf.Material, surf.ExtraStages, tu, tv, cp.ProjectZ(1.0))
			cp.SampleLight(&cp.MaterialSampler.Output, surf.Material, &cp.RaySegIntersect, cp.Distance)
		} else {
			cp.MaterialSampler.Output[0] = 0.5
			cp.MaterialSampler.Output[1] = 1
			cp.MaterialSampler.Output[2] = 0.5
			cp.MaterialSampler.Output[3] = 1
		}
		cp.FrameBuffer[screenIndex].AddPreMulColorSelf(&cp.MaterialSampler.Output)
		cp.ZBuffer[screenIndex] = cp.Distance
	}
}

func WallLowPick(cp *state.ColumnPortal) {
	if cp.ScreenY >= cp.AdjClippedBottom && cp.ScreenY < cp.ClippedBottom {
		cp.PickedSelection = append(cp.PickedSelection, core.SelectableFromWall(cp.AdjSegment, core.SelectableLow))
	}
}

// WallLow renders the bottom portion of a portal segment.
func WallLow(cp *state.ColumnPortal) {
	surf := cp.AdjSegment.LoSurface
	transform := surf.Transform.Render
	// To calculate the vertical texture coordinate, we can't use the integer
	// screen coordinates, we need to use the precise floats
	vStart := float64(cp.ScreenHeight/2) - cp.AdjProjectedBottom
	for cp.ScreenY = cp.AdjClippedBottom; cp.ScreenY < cp.ClippedBottom; cp.ScreenY++ {
		screenIndex := uint32(cp.ScreenX + cp.ScreenY*cp.ScreenWidth)
		if cp.Distance >= cp.ZBuffer[screenIndex] {
			continue
		}
		v := (float64(cp.ScreenY) - vStart) / (cp.AdjProjectedBottom - cp.ProjectedBottom)
		cp.RaySegIntersect[2] = (1.0-v)*cp.AdjBottom + v*cp.IntersectionBottom

		if surf.Material != 0 {
			tu := transform[0]*cp.U + transform[2]*v + transform[4]
			tv := transform[1]*cp.U + transform[3]*v + transform[5]
			cp.SampleShader(surf.Material, surf.ExtraStages, tu, tv, cp.ProjectZ(1.0))
			cp.SampleLight(&cp.MaterialSampler.Output, surf.Material, &cp.RaySegIntersect, cp.Distance)
		} else {
			cp.MaterialSampler.Output[0] = 0.5
			cp.MaterialSampler.Output[1] = 1
			cp.MaterialSampler.Output[2] = 0.5
			cp.MaterialSampler.Output[3] = 1
		}
		cp.FrameBuffer[screenIndex].AddPreMulColorSelf(&cp.MaterialSampler.Output)
		cp.ZBuffer[screenIndex] = cp.Distance
	}
}
