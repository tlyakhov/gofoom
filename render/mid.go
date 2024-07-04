// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package render

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/materials"
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
	u := c.U
	if surf.Stretch == materials.StretchNone {
		if !internalSegment && c.Sector.Winding < 0 {
			u = 1.0 - u
		}
		u = (c.Segment.A[0] + c.Segment.A[1] + u*c.Segment.Length) / 64.0
	}
	for c.ScreenY = c.ClippedStart; c.ScreenY < c.ClippedEnd; c.ScreenY++ {
		screenIndex := uint32(c.ScreenX + c.ScreenY*c.ScreenWidth)

		if c.Distance >= c.ZBuffer[screenIndex] {
			continue
		}
		v := float64(c.ScreenY-c.ScreenStart) / float64(c.ScreenEnd-c.ScreenStart)
		c.RaySegIntersect[2] = c.TopZ + v*(c.BottomZ-c.TopZ)

		if surf.Stretch == materials.StretchNone {
			v = -c.RaySegIntersect[2] / 64.0
		} else if surf.Stretch == materials.StretchAspect {
			v *= (c.Segment.Top - c.Segment.Bottom) / c.Segment.Length
		}

		if surf.Material != 0 {
			tu := surf.Transform[0]*u + surf.Transform[2]*v + surf.Transform[4]
			tv := surf.Transform[1]*u + surf.Transform[3]*v + surf.Transform[5]
			c.SampleShader(surf.Material, surf.ExtraStages, tu, tv, c.ProjectZ(1.0))
			c.SampleLight(&c.MaterialSampler.Output, surf.Material, &c.RaySegIntersect, c.Distance)
		}
		c.ApplySample(&c.MaterialSampler.Output, int(screenIndex), c.Distance)
	}
}
