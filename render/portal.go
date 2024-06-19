// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package render

import (
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/render/state"
)

func WallHiPick(s *state.ColumnPortal) {
	if s.ScreenY >= s.ClippedStart && s.ScreenY < s.AdjClippedTop {
		s.PickedElements = append(s.PickedElements, state.PickedElement{Type: state.PickHigh, Element: s.AdjSegment})
	}
}

// WallHi renders the top portion of a portal segment.
func WallHi(s *state.ColumnPortal) {
	surf := s.AdjSegment.HiSurface
	u := s.U
	if surf.Stretch == materials.StretchNone {
		if s.Sector.Winding < 0 {
			u = 1.0 - u
		}
		u = (s.Segment.A[0] + s.Segment.A[1] + u*s.Segment.Length) / 64.0
	}
	for s.ScreenY = s.ClippedStart; s.ScreenY < s.AdjClippedTop; s.ScreenY++ {
		screenIndex := uint32(s.ScreenX + s.ScreenY*s.ScreenWidth)
		if s.Distance >= s.ZBuffer[screenIndex] {
			continue
		}
		v := float64(s.ScreenY-s.ScreenStart) / float64(s.AdjScreenTop-s.ScreenStart)
		s.RaySegIntersect[2] = (1.0-v)*s.TopZ + v*s.AdjCeilZ

		if surf.Stretch == materials.StretchNone {
			v = -s.RaySegIntersect[2] / 64.0
		} else if surf.Stretch == materials.StretchAspect {
			v *= (s.Sector.Max[2] - s.Sector.Min[2]) / s.Segment.Length
		}

		if !surf.Material.Nil() {
			tu := surf.Transform[0]*u + surf.Transform[2]*v + surf.Transform[4]
			tv := surf.Transform[1]*u + surf.Transform[3]*v + surf.Transform[5]
			s.SampleShader(surf.Material, surf.ExtraStages, tu, tv, s.ProjectZ(1.0))
			s.SampleLight(&s.MaterialSampler.Output, surf.Material, &s.RaySegIntersect, s.Distance)
		}
		//concepts.AsmVector4AddPreMulColorSelf((*[4]float64)(&s.FrameBuffer[screenIndex]), (*[4]float64)(&s.Material))
		s.FrameBuffer[screenIndex].AddPreMulColorSelf(&s.MaterialSampler.Output)
		s.ZBuffer[screenIndex] = s.Distance
	}
}

func WallLowPick(s *state.ColumnPortal) {
	if s.ScreenY >= s.AdjClippedBottom && s.ScreenY < s.ClippedEnd {
		s.PickedElements = append(s.PickedElements, state.PickedElement{Type: state.PickLow, Element: s.AdjSegment})
	}
}

// WallLow renders the bottom portion of a portal segment.
func WallLow(s *state.ColumnPortal) {
	surf := s.AdjSegment.LoSurface
	u := s.U
	if surf.Stretch == materials.StretchNone {
		if s.Sector.Winding < 0 {
			u = 1.0 - u
		}
		u = (s.Segment.A[0] + s.Segment.A[1] + u*s.Segment.Length) / 64.0
	}
	for s.ScreenY = s.AdjClippedBottom; s.ScreenY < s.ClippedEnd; s.ScreenY++ {
		screenIndex := uint32(s.ScreenX + s.ScreenY*s.ScreenWidth)
		if s.Distance >= s.ZBuffer[screenIndex] {
			continue
		}
		v := float64(s.ScreenY-s.AdjScreenBottom) / float64(s.ScreenEnd-s.AdjScreenBottom)
		s.RaySegIntersect[2] = (1.0-v)*s.AdjFloorZ + v*s.BottomZ

		if surf.Stretch == materials.StretchNone {
			v = -s.RaySegIntersect[2] / 64.0
		} else if surf.Stretch == materials.StretchAspect {
			v *= (s.Sector.Max[2] - s.Sector.Min[2]) / s.Segment.Length
		}

		if !surf.Material.Nil() {
			tu := surf.Transform[0]*u + surf.Transform[2]*v + surf.Transform[4]
			tv := surf.Transform[1]*u + surf.Transform[3]*v + surf.Transform[5]
			s.SampleShader(surf.Material, surf.ExtraStages, tu, tv, s.ProjectZ(1.0))
			s.SampleLight(&s.MaterialSampler.Output, surf.Material, &s.RaySegIntersect, s.Distance)
		}
		//concepts.AsmVector4AddPreMulColorSelf((*[4]float64)(&s.FrameBuffer[screenIndex]), (*[4]float64)(&s.Material))
		s.FrameBuffer[screenIndex].AddPreMulColorSelf(&s.MaterialSampler.Output)
		s.ZBuffer[screenIndex] = s.Distance
	}
}
