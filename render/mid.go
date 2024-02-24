package render

import (
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/render/state"
)

func WallMidPick(s *state.Column) {
	if s.Y >= s.ClippedStart && s.Y < s.ClippedEnd {
		s.PickedElements = append(s.PickedElements, state.PickedElement{Type: state.PickMid, Element: s.Segment})
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
	for s.Y = s.ClippedStart; s.Y < s.ClippedEnd; s.Y++ {
		screenIndex := uint32(s.X + s.Y*s.ScreenWidth)

		if s.Distance >= s.ZBuffer[screenIndex] {
			continue
		}
		v := float64(s.Y-s.ScreenStart) / float64(s.ScreenEnd-s.ScreenStart)
		s.Intersection[2] = s.TopZ + v*(s.BottomZ-s.TopZ)
		lightV := v
		if !internalSegment && (s.Sector.FloorSlope != 0 || s.Sector.CeilSlope != 0) {
			lightV = (s.Sector.Max[2] - s.Intersection[2]) / (s.Sector.Max[2] - s.Sector.Min[2])
		}

		if surf.Stretch == materials.StretchNone {
			v = -s.Intersection[2] / 64.0
		} else if surf.Stretch == materials.StretchAspect {
			v *= (s.Segment.Top - s.Segment.Bottom) / s.Segment.Length
		}

		if !surf.Material.Nil() {
			tu := surf.Transform[0]*u + surf.Transform[2]*v + surf.Transform[4]
			tv := surf.Transform[1]*u + surf.Transform[3]*v + surf.Transform[5]
			s.SampleShader(surf.Material, surf.ExtraStages, tu, tv, s.ProjectZ(1.0))
			s.SampleLight(&s.MaterialSampler.Output, surf.Material, &s.Intersection, s.U, lightV, s.Distance)
		}
		s.ApplySample(&s.MaterialSampler.Output, int(screenIndex), s.Distance)
	}
}
