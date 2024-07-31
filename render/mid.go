// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package render

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/render/state"
)

func wallPick(s *state.Column) {
	if s.ScreenY >= s.ClippedTop && s.ScreenY < s.ClippedBottom {
		s.PickedSelection = append(s.PickedSelection, core.SelectableFromWall(s.SectorSegment, core.SelectableMid))
	}
}

// wall renders the wall portion (potentially over a portal).
func (r *Renderer) wall(c *state.Column) {
	surf := c.Segment.Surface
	transform := surf.Transform.Render
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
	sw := uint32(c.ProjectZ(c.Segment.Length))
	sh := uint32(dv)
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

		if surf.Material != 0 {
			tu := transform[0]*c.U + transform[2]*v + transform[4]
			tv := transform[1]*c.U + transform[3]*v + transform[5]
			//log.Printf("%v,%v", sw, sh)
			c.SampleShader(surf.Material, surf.ExtraStages, tu, tv, sw, sh)
			c.SampleLight(&c.MaterialSampler.Output, surf.Material, &c.RaySegIntersect, c.Distance)
		}
		r.ApplySample(&c.MaterialSampler.Output, int(screenIndex), c.Distance)
	}
}
