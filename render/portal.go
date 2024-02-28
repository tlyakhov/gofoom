package render

import (
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/render/state"
)

func WallHiPick(s *state.ColumnPortal) {
	if s.Y >= s.ClippedStart && s.Y < s.AdjClippedTop {
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
	for s.Y = s.ClippedStart; s.Y < s.AdjClippedTop; s.Y++ {
		screenIndex := uint32(s.X + s.Y*s.ScreenWidth)
		if s.Distance >= s.ZBuffer[screenIndex] {
			continue
		}
		v := float64(s.Y-s.ScreenStart) / float64(s.AdjScreenTop-s.ScreenStart)
		s.RaySegIntersect[2] = (1.0-v)*s.TopZ + v*s.AdjCeilZ
		lightV := (s.Sector.Max[2] - s.RaySegIntersect[2]) / (s.Sector.Max[2] - s.Sector.Min[2])

		if surf.Stretch == materials.StretchNone {
			v = -s.RaySegIntersect[2] / 64.0
		} else if surf.Stretch == materials.StretchAspect {
			v *= (s.Sector.Max[2] - s.Sector.Min[2]) / s.Segment.Length
		}

		if !surf.Material.Nil() {
			tu := surf.Transform[0]*u + surf.Transform[2]*v + surf.Transform[4]
			tv := surf.Transform[1]*u + surf.Transform[3]*v + surf.Transform[5]
			s.SampleShader(surf.Material, surf.ExtraStages, tu, tv, s.ProjectZ(1.0))
			s.SampleLight(&s.MaterialSampler.Output, surf.Material, &s.RaySegIntersect, s.U, lightV, s.Distance)
		}
		//concepts.AsmVector4AddPreMulColorSelf((*[4]float64)(&s.FrameBuffer[screenIndex]), (*[4]float64)(&s.Material))
		s.FrameBuffer[screenIndex].AddPreMulColorSelf(&s.MaterialSampler.Output)
		s.ZBuffer[screenIndex] = s.Distance
	}
}

func WallLowPick(s *state.ColumnPortal) {
	if s.Y >= s.AdjClippedBottom && s.Y < s.ClippedEnd {
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
	for s.Y = s.AdjClippedBottom; s.Y < s.ClippedEnd; s.Y++ {
		screenIndex := uint32(s.X + s.Y*s.ScreenWidth)
		if s.Distance >= s.ZBuffer[screenIndex] {
			continue
		}
		v := float64(s.Y-s.AdjScreenBottom) / float64(s.ScreenEnd-s.AdjScreenBottom)
		s.RaySegIntersect[2] = (1.0-v)*s.AdjFloorZ + v*s.BottomZ
		lightV := (s.Sector.Max[2] - s.RaySegIntersect[2]) / (s.Sector.Max[2] - s.Sector.Min[2])

		if surf.Stretch == materials.StretchNone {
			v = -s.RaySegIntersect[2] / 64.0
		} else if surf.Stretch == materials.StretchAspect {
			v *= (s.Sector.Max[2] - s.Sector.Min[2]) / s.Segment.Length
		}

		if !surf.Material.Nil() {
			tu := surf.Transform[0]*u + surf.Transform[2]*v + surf.Transform[4]
			tv := surf.Transform[1]*u + surf.Transform[3]*v + surf.Transform[5]
			s.SampleShader(surf.Material, surf.ExtraStages, tu, tv, s.ProjectZ(1.0))
			s.SampleLight(&s.MaterialSampler.Output, surf.Material, &s.RaySegIntersect, s.U, lightV, s.Distance)
		}
		//concepts.AsmVector4AddPreMulColorSelf((*[4]float64)(&s.FrameBuffer[screenIndex]), (*[4]float64)(&s.Material))
		s.FrameBuffer[screenIndex].AddPreMulColorSelf(&s.MaterialSampler.Output)
		s.ZBuffer[screenIndex] = s.Distance
	}
}
