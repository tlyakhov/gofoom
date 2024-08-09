// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package render

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/render/state"
)

func wallHiPick(cp *state.ColumnPortal) {
	if cp.ScreenY >= cp.ClippedTop && cp.ScreenY < cp.AdjClippedTop {
		cp.PickedSelection = append(cp.PickedSelection, core.SelectableFromWall(cp.AdjSegment, core.SelectableHi))
	}
}

// wallHi renders the top portion of a portal segment.
func wallHi(cp *state.ColumnPortal) {
	mat := cp.AdjSegment.HiSurface.Material
	extras := cp.AdjSegment.HiSurface.ExtraStages
	cp.MaterialSampler.Initialize(mat, extras)
	transform := cp.AdjSegment.HiSurface.Transform.Render
	cp.ScaleW = uint32(cp.ProjectZ(cp.SectorSegment.Segment.Length))
	cp.ScaleH = uint32(cp.ProjectedTop - cp.AdjProjectedTop)
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

		if mat != 0 {
			cp.MaterialSampler.NU = cp.SegmentIntersection.U
			cp.MaterialSampler.NV = v
			cp.MaterialSampler.U = transform[0]*cp.MaterialSampler.NU + transform[2]*cp.MaterialSampler.NV + transform[4]
			cp.MaterialSampler.V = transform[1]*cp.MaterialSampler.NU + transform[3]*cp.MaterialSampler.NV + transform[5]
			cp.SampleMaterial(extras)
			cp.SampleLight(&cp.MaterialSampler.Output, mat, &cp.RaySegIntersect, cp.Distance)
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

func wallLowPick(cp *state.ColumnPortal) {
	if cp.ScreenY >= cp.AdjClippedBottom && cp.ScreenY < cp.ClippedBottom {
		cp.PickedSelection = append(cp.PickedSelection, core.SelectableFromWall(cp.AdjSegment, core.SelectableLow))
	}
}

// wallLow renders the bottom portion of a portal segment.
func wallLow(cp *state.ColumnPortal) {
	mat := cp.AdjSegment.LoSurface.Material
	extras := cp.AdjSegment.LoSurface.ExtraStages
	cp.MaterialSampler.Initialize(mat, extras)
	transform := cp.AdjSegment.LoSurface.Transform.Render
	cp.ScaleW = uint32(cp.ProjectZ(cp.SectorSegment.Segment.Length))
	cp.ScaleH = uint32(cp.AdjProjectedBottom - cp.ProjectedBottom)
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

		if mat != 0 {
			cp.MaterialSampler.NU = cp.SegmentIntersection.U
			cp.MaterialSampler.NV = v
			cp.MaterialSampler.U = transform[0]*cp.MaterialSampler.NU + transform[2]*cp.MaterialSampler.NV + transform[4]
			cp.MaterialSampler.V = transform[1]*cp.MaterialSampler.NU + transform[3]*cp.MaterialSampler.NV + transform[5]
			cp.SampleMaterial(extras)
			cp.SampleLight(&cp.MaterialSampler.Output, mat, &cp.RaySegIntersect, cp.Distance)
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
