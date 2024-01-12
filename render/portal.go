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
	mat := s.AdjSegment.HiSurface.Material
	extras := s.AdjSegment.HiSurface.ExtraStages
	transform := s.AdjSegment.HiSurface.Transform
	u := s.U
	if s.Segment.HiSurface.Scale == materials.ScaleHeight || s.Segment.HiSurface.Scale == materials.ScaleNone {
		if s.Sector.Winding < 0 {
			u = 1.0 - u
		}
		u = (s.Segment.P[0] + s.Segment.P[1] + u*s.Segment.Length) / 64.0
	}
	for s.Y = s.ClippedStart; s.Y < s.AdjClippedTop; s.Y++ {
		screenIndex := uint32(s.X + s.Y*s.ScreenWidth)
		if s.Distance >= s.ZBuffer[screenIndex] {
			continue
		}
		v := float64(s.Y-s.ScreenStart) / float64(s.AdjScreenTop-s.ScreenStart)
		s.Intersection[2] = (1.0-v)*s.CeilZ + v*s.AdjCeilZ
		lightV := (s.Sector.Max[2] - s.Intersection[2]) / (s.Sector.Max[2] - s.Sector.Min[2])

		if s.Segment.HiSurface.Scale == materials.ScaleWidth || s.Segment.HiSurface.Scale == materials.ScaleNone {
			v = s.Intersection[2] / 64.0
		}

		if !mat.Nil() {
			tu := transform[0]*u + transform[2]*v + transform[4]
			tv := transform[1]*u + transform[3]*v + transform[5]
			s.SampleShader(mat, extras, tu, tv, s.ProjectZ(1.0))
			s.SampleLight(&s.Material, mat, &s.Intersection, s.U, lightV, s.Distance)
		}
		//concepts.AsmVector4AddPreMulColorSelf((*[4]float64)(&s.FrameBuffer[screenIndex]), (*[4]float64)(&s.Material))
		s.FrameBuffer[screenIndex].AddPreMulColorSelf(&s.Material)
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
	mat := s.AdjSegment.LoSurface.Material
	extras := s.AdjSegment.LoSurface.ExtraStages
	transform := s.AdjSegment.LoSurface.Transform
	u := s.U
	if s.Segment.LoSurface.Scale == materials.ScaleHeight || s.Segment.LoSurface.Scale == materials.ScaleNone {
		if s.Sector.Winding < 0 {
			u = 1.0 - u
		}
		u = (s.Segment.P[0] + s.Segment.P[1] + u*s.Segment.Length) / 64.0
	}
	for s.Y = s.AdjClippedBottom; s.Y < s.ClippedEnd; s.Y++ {
		screenIndex := uint32(s.X + s.Y*s.ScreenWidth)
		if s.Distance >= s.ZBuffer[screenIndex] {
			continue
		}
		v := float64(s.Y-s.AdjScreenBottom) / float64(s.ScreenEnd-s.AdjScreenBottom)
		s.Intersection[2] = (1.0-v)*s.AdjFloorZ + v*s.FloorZ
		lightV := (s.Sector.Max[2] - s.Intersection[2]) / (s.Sector.Max[2] - s.Sector.Min[2])

		if s.Segment.LoSurface.Scale == materials.ScaleWidth || s.Segment.LoSurface.Scale == materials.ScaleNone {
			v = s.Intersection[2] / 64.0
		}

		if !mat.Nil() {
			tu := transform[0]*u + transform[2]*v + transform[4]
			tv := transform[1]*u + transform[3]*v + transform[5]
			s.SampleShader(mat, extras, tu, tv, s.ProjectZ(1.0))
			s.SampleLight(&s.Material, mat, &s.Intersection, s.U, lightV, s.Distance)
		}
		//concepts.AsmVector4AddPreMulColorSelf((*[4]float64)(&s.FrameBuffer[screenIndex]), (*[4]float64)(&s.Material))
		s.FrameBuffer[screenIndex].AddPreMulColorSelf(&s.Material)
		s.ZBuffer[screenIndex] = s.Distance
	}
}
