// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package render

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/render/state"
)

func WallMidPick(s *state.Column) {
	if s.ScreenY >= s.ClippedStart && s.ScreenY < s.ClippedEnd {
		s.PickedSelection = append(s.PickedSelection, core.SelectableFromWall(s.SectorSegment, core.SelectableMid))
	}
}

// WallMid renders the wall portion (potentially over a portal).
func WallMid(c *state.Column, internalSegment bool) {
	surf := c.Segment.Surface
	transform := surf.Transform.Render
	for c.ScreenY = c.ClippedStart; c.ScreenY < c.ClippedEnd; c.ScreenY++ {
		screenIndex := uint32(c.ScreenX + c.ScreenY*c.ScreenWidth)

		if c.Distance >= c.ZBuffer[screenIndex] {
			continue
		}
		v := float64(c.ScreenY-c.ScreenStart) / float64(c.ScreenEnd-c.ScreenStart)
		c.RaySegIntersect[2] = c.TopZ + v*(c.BottomZ-c.TopZ)

		if surf.Material != 0 {
			tu := transform[0]*c.U + transform[2]*v + transform[4]
			tv := transform[1]*c.U + transform[3]*v + transform[5]
			c.SampleShader(surf.Material, surf.ExtraStages, tu, tv, c.ProjectZ(1.0))
			c.SampleLight(&c.MaterialSampler.Output, surf.Material, &c.RaySegIntersect, c.Distance)
		}
		c.ApplySample(&c.MaterialSampler.Output, int(screenIndex), c.Distance)
	}
}
