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
func WallMid(s *state.Column, internalSegment bool) {
	surf := s.Segment.Surface
	u := s.U
	if surf.Stretch == materials.StretchNone {
		if !internalSegment && s.Sector.Winding < 0 {
			u = 1.0 - u
		}
		u = (s.Segment.A[0] + s.Segment.A[1] + u*s.Segment.Length) / 64.0
	}
	for s.ScreenY = s.ClippedStart; s.ScreenY < s.ClippedEnd; s.ScreenY++ {
		screenIndex := uint32(s.ScreenX + s.ScreenY*s.ScreenWidth)

		if s.Distance >= s.ZBuffer[screenIndex] {
			continue
		}
		v := float64(s.ScreenY-s.ScreenStart) / float64(s.ScreenEnd-s.ScreenStart)
		s.RaySegIntersect[2] = s.TopZ + v*(s.BottomZ-s.TopZ)

		if surf.Stretch == materials.StretchNone {
			v = -s.RaySegIntersect[2] / 64.0
		} else if surf.Stretch == materials.StretchAspect {
			v *= (s.Segment.Top - s.Segment.Bottom) / s.Segment.Length
		}

		if !surf.Material.Nil() {
			tu := surf.Transform[0]*u + surf.Transform[2]*v + surf.Transform[4]
			tv := surf.Transform[1]*u + surf.Transform[3]*v + surf.Transform[5]
			s.SampleShader(surf.Material, surf.ExtraStages, tu, tv, s.ProjectZ(1.0))
			s.SampleLight(&s.MaterialSampler.Output, surf.Material, &s.RaySegIntersect, s.Distance)
		}
		s.ApplySample(&s.MaterialSampler.Output, int(screenIndex), s.Distance)
	}
}
