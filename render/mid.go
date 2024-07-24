// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package render

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/render/state"
)

func WallMidPick(s *state.Column) {
	if s.ScreenY >= s.ClippedTop && s.ScreenY < s.ClippedBottom {
		s.PickedSelection = append(s.PickedSelection, core.SelectableFromWall(s.SectorSegment, core.SelectableMid))
	}
}

// WallMid renders the wall portion (potentially over a portal).
func WallMid(c *state.Column, internalSegment bool) {
	surf := c.Segment.Surface
	transform := surf.Transform.Render
	// To calculate the vertical texture coordinate, we can't use the integer
	// screen coordinates, we need to use the precise floats
	vStart := float64(c.ScreenHeight/2) - c.ProjectedTop
	for c.ScreenY = c.ClippedTop; c.ScreenY < c.ClippedBottom; c.ScreenY++ {
		screenIndex := uint32(c.ScreenX + c.ScreenY*c.ScreenWidth)

		if c.Distance >= c.ZBuffer[screenIndex] {
			continue
		}
		v := (float64(c.ScreenY) - vStart) / (c.ProjectedTop - c.ProjectedBottom)
		c.RaySegIntersect[2] = c.IntersectionTop*(1.0-v) + v*c.IntersectionBottom

		if surf.Material != 0 {
			tu := transform[0]*c.U + transform[2]*v + transform[4]
			tv := transform[1]*c.U + transform[3]*v + transform[5]
			c.SampleShader(surf.Material, surf.ExtraStages, tu, tv, c.ProjectZ(1.0))
			c.SampleLight(&c.MaterialSampler.Output, surf.Material, &c.RaySegIntersect, c.Distance)
		}
		c.ApplySample(&c.MaterialSampler.Output, int(screenIndex), c.Distance)
	}
}
