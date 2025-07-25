// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package render

import (
	"math"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/components/selection"
	"tlyakhov/gofoom/concepts"
)

func wallHiPick(cp *columnPortal) {
	if cp.ScreenY >= cp.ClippedTop && cp.ScreenY < cp.AdjClippedTop {
		cp.PickedSelection = append(cp.PickedSelection, selection.SelectableFromWall(cp.AdjSegment, selection.SelectableHi))
	}
}

// wallHi renders the top portion of a portal segment.
func wallHi(cp *columnPortal) {
	mat := cp.AdjSegment.HiSurface.Material
	lit := materials.GetLit(cp.AdjSegment.HiSurface.Material)
	extras := cp.AdjSegment.HiSurface.ExtraStages
	cp.MaterialSampler.Initialize(mat, extras)
	transform := cp.AdjSegment.HiSurface.Transform.Render
	cp.ScaleW = uint32(cp.ProjectZ(cp.SectorSegment.Segment.Length))
	cp.ScaleH = uint32(cp.ProjectedTop - cp.AdjProjectedTop)
	// To calculate the vertical texture coordinate, we can't use the integer
	// screen coordinates, we need to use the precise floats
	vStart := float64(cp.ScreenHeight/2) - cp.ProjectedTop + math.Floor(cp.ShearZ)
	for cp.ScreenY = cp.ClippedTop; cp.ScreenY < cp.AdjClippedTop; cp.ScreenY++ {
		screenIndex := uint32(cp.ScreenX + cp.ScreenY*cp.ScreenWidth)
		if cp.Distance >= cp.ZBuffer[screenIndex] {
			continue
		}
		v := (float64(cp.ScreenY) - vStart) / (cp.ProjectedTop - cp.AdjProjectedTop)
		cp.RaySegIntersect[2] = (1.0-v)*cp.IntersectionTop + v*cp.AdjTop

		if mat != 0 {
			cp.MaterialSampler.NU = cp.segmentIntersection.U
			cp.MaterialSampler.NV = v
			cp.MaterialSampler.U = transform[0]*cp.MaterialSampler.NU + transform[2]*cp.MaterialSampler.NV + transform[4]
			cp.MaterialSampler.V = transform[1]*cp.MaterialSampler.NU + transform[3]*cp.MaterialSampler.NV + transform[5]
			cp.SampleMaterial(extras)
			if lit != nil {
				cp.SampleLight(&cp.MaterialSampler.Output, lit, &cp.RaySegIntersect, cp.Distance)
			}
		} else {
			cp.MaterialSampler.Output[0] = 0.5
			cp.MaterialSampler.Output[1] = 1
			cp.MaterialSampler.Output[2] = 0.5
			cp.MaterialSampler.Output[3] = 1
		}
		concepts.BlendColors(&cp.FrameBuffer[screenIndex], &cp.MaterialSampler.Output, 1)
		cp.ZBuffer[screenIndex] = cp.Distance
	}
}

func wallLowPick(cp *columnPortal) {
	if cp.ScreenY >= cp.AdjClippedBottom && cp.ScreenY < cp.ClippedBottom {
		cp.PickedSelection = append(cp.PickedSelection, selection.SelectableFromWall(cp.AdjSegment, selection.SelectableLow))
	}
}

// wallLow renders the bottom portion of a portal segment.
func wallLow(cp *columnPortal) {
	mat := cp.AdjSegment.LoSurface.Material
	lit := materials.GetLit(cp.AdjSegment.LoSurface.Material)
	extras := cp.AdjSegment.LoSurface.ExtraStages
	cp.MaterialSampler.Initialize(mat, extras)
	transform := cp.AdjSegment.LoSurface.Transform.Render
	cp.ScaleW = uint32(cp.ProjectZ(cp.SectorSegment.Segment.Length))
	cp.ScaleH = uint32(cp.AdjProjectedBottom - cp.ProjectedBottom)
	// To calculate the vertical texture coordinate, we can't use the integer
	// screen coordinates, we need to use the precise floats
	vStart := float64(cp.ScreenHeight/2) - cp.AdjProjectedBottom + math.Floor(cp.ShearZ)
	for cp.ScreenY = cp.AdjClippedBottom; cp.ScreenY < cp.ClippedBottom; cp.ScreenY++ {
		screenIndex := uint32(cp.ScreenX + cp.ScreenY*cp.ScreenWidth)
		if cp.Distance >= cp.ZBuffer[screenIndex] {
			continue
		}
		v := (float64(cp.ScreenY) - vStart) / (cp.AdjProjectedBottom - cp.ProjectedBottom)
		cp.RaySegIntersect[2] = (1.0-v)*cp.AdjBottom + v*cp.IntersectionBottom

		if mat != 0 {
			cp.MaterialSampler.NU = cp.segmentIntersection.U
			cp.MaterialSampler.NV = v
			cp.MaterialSampler.U = transform[0]*cp.MaterialSampler.NU + transform[2]*cp.MaterialSampler.NV + transform[4]
			cp.MaterialSampler.V = transform[1]*cp.MaterialSampler.NU + transform[3]*cp.MaterialSampler.NV + transform[5]
			cp.SampleMaterial(extras)
			if lit != nil {
				cp.SampleLight(&cp.MaterialSampler.Output, lit, &cp.RaySegIntersect, cp.Distance)
			}
		} else {
			cp.MaterialSampler.Output[0] = 0.5
			cp.MaterialSampler.Output[1] = 1
			cp.MaterialSampler.Output[2] = 0.5
			cp.MaterialSampler.Output[3] = 1
		}
		concepts.BlendColors(&cp.FrameBuffer[screenIndex], &cp.MaterialSampler.Output, 1)
		cp.ZBuffer[screenIndex] = cp.Distance
	}
}
