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
	mat := s.Segment.MidSurface.Material
	u := s.U
	if s.Segment.MidSurface.Scale == materials.ScaleHeight || s.Segment.MidSurface.Scale == materials.ScaleNone {
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

		if s.Segment.MidSurface.Scale == materials.ScaleWidth || s.Segment.MidSurface.Scale == materials.ScaleNone {
			v = s.Intersection[2] / 64.0
		}

		if !mat.Nil() {
			s.SampleShader(mat, u, v, s.ProjectZ(1.0))
			s.SampleLight(&s.Material, mat, &s.Intersection, s.U, lightV, s.Distance)
		}
		//concepts.AsmVector4AddPreMulColorSelf((*[4]float64)(&s.FrameBuffer[screenIndex]), (*[4]float64)(&s.Material))
		s.FrameBuffer[screenIndex].AddPreMulColorSelf(&s.Material)
		s.ZBuffer[screenIndex] = s.Distance
	}
}
