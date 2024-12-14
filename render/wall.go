// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package render

import (
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/components/selection"
	"tlyakhov/gofoom/concepts"
)

func wallPick(s *column) {
	if s.ScreenY >= s.ClippedTop && s.ScreenY < s.ClippedBottom {
		s.PickedSelection = append(s.PickedSelection, selection.SelectableFromWall(s.SectorSegment, selection.SelectableMid))
	}
}

// wall renders the wall portion (potentially over a portal).
func (r *Renderer) wall(c *column) {
	mat := c.Segment.Surface.Material
	lit := materials.GetLit(c.ECS, c.Segment.Surface.Material)
	extras := c.Segment.Surface.ExtraStages
	c.MaterialSampler.Initialize(mat, extras)
	transform := c.Segment.Surface.Transform.Render
	noSlope := c.SectorSegment != nil && c.SectorSegment.WallUVIgnoreSlope
	// To calculate the vertical texture coordinate, we can't use the integer
	// screen coordinates, we need to use the precise floats
	dv := 0.0
	vTop := float64(c.ScreenHeight / 2)
	if noSlope {
		vTop -= c.ProjectedSectorTop
		dv = c.ProjectedSectorTop - c.ProjectedSectorBottom
	} else {
		vTop -= c.ProjectedTop
		dv = (c.ProjectedTop - c.ProjectedBottom)
	}
	c.ScaleW = uint32(c.ProjectZ(c.Segment.Length))
	c.ScaleH = uint32(dv)
	if dv != 0 {
		dv = 1.0 / dv
	}

	for c.ScreenY = c.ClippedTop; c.ScreenY < c.ClippedBottom; c.ScreenY++ {
		screenIndex := uint32(c.ScreenX + c.ScreenY*c.ScreenWidth)

		if c.Distance >= c.ZBuffer[screenIndex] {
			continue
		}

		v := (float64(c.ScreenY) - vTop) * dv
		c.RaySegIntersect[2] = c.IntersectionTop*(1.0-v) + v*c.IntersectionBottom

		if mat != 0 {
			c.MaterialSampler.NU = c.segmentIntersection.U
			c.MaterialSampler.NV = v
			c.MaterialSampler.U = transform[0]*c.MaterialSampler.NU + transform[2]*c.MaterialSampler.NV + transform[4]
			c.MaterialSampler.V = transform[1]*c.MaterialSampler.NU + transform[3]*c.MaterialSampler.NV + transform[5]
			c.SampleMaterial(extras)
			if lit != nil {
				c.SampleLight(&c.MaterialSampler.Output, lit, &c.RaySegIntersect, c.Distance)
			}
		}
		concepts.BlendColors((*[4]float64)(&r.FrameBuffer[screenIndex]), (*[4]float64)(&c.MaterialSampler.Output), 1.0)
		if c.MaterialSampler.Output[3] > 0.8 {
			r.ZBuffer[screenIndex] = c.Distance
		}
	}
}
