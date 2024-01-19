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
func WallMid(s *state.Column) {
	material := s.Segment.MidSurface.Material
	extras := s.Segment.MidSurface.ExtraStages
	transform := s.Segment.MidSurface.Transform
	u := s.U
	if s.Segment.MidSurface.Stretch == materials.StretchNone {
		if s.Sector.Winding < 0 {
			u = 1.0 - u
		}
		u = (s.Segment.P[0] + s.Segment.P[1] + u*s.Segment.Length) / 64.0
	}
	for s.Y = s.ClippedStart; s.Y < s.ClippedEnd; s.Y++ {
		screenIndex := uint32(s.X + s.Y*s.ScreenWidth)

		if s.Distance >= s.ZBuffer[screenIndex] {
			continue
		}
		v := float64(s.Y-s.ScreenStart) / float64(s.ScreenEnd-s.ScreenStart)
		s.Intersection[2] = s.CeilZ + v*(s.FloorZ-s.CeilZ)
		lightV := v
		if s.Sector.FloorSlope != 0 || s.Sector.CeilSlope != 0 {
			lightV = (s.Sector.Max[2] - s.Intersection[2]) / (s.Sector.Max[2] - s.Sector.Min[2])
		}

		if s.Segment.MidSurface.Stretch == materials.StretchNone {
			v = s.Intersection[2] / 64.0
		} else if s.Segment.MidSurface.Stretch == materials.StretchAspect {
			v *= (s.Sector.Max[2] - s.Sector.Min[2]) / s.Segment.Length
		}

		if !material.Nil() {
			tu := transform[0]*u + transform[2]*v + transform[4]
			tv := transform[1]*u + transform[3]*v + transform[5]
			s.SampleShader(material, extras, tu, tv, s.ProjectZ(1.0))
			s.SampleLight(&s.MaterialColor, material, &s.Intersection, s.U, lightV, s.Distance)
		}
		s.ApplySample(&s.MaterialColor, int(screenIndex), s.Distance)
	}
}
