package render

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/render/state"
)

func WallMidPick(s *state.Slice) {
	if s.Y >= s.ClippedStart && s.Y < s.ClippedEnd {
		s.PickedElements = append(s.PickedElements, state.PickedElement{Type: state.PickMid, Element: s.Segment})
	}
}

// WallMid renders the wall portion (potentially over a portal).
func WallMid(s *state.Slice) {
	mat := s.Segment.MidMaterial
	u := s.U
	if s.Segment.MidBehavior == core.ScaleHeight || s.Segment.MidBehavior == core.ScaleNone {
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
		s.Intersection[2] = (1.0-v)*s.CeilZ + v*s.FloorZ
		lightV := v
		if s.Sector.FloorSlope != 0 || s.Sector.CeilSlope != 0 {
			lightV = (s.Sector.Max[2] - s.Intersection[2]) / (s.Sector.Max[2] - s.Sector.Min[2])
		}

		if s.Segment.MidBehavior == core.ScaleWidth || s.Segment.MidBehavior == core.ScaleNone {
			v = s.Intersection[2] / 64.0
		}

		c := &concepts.Vector4{0.5, 0.5, 0.5, 1}
		if !mat.Nil() {
			*c = s.SampleMaterial(mat, u, v, s.ProjectZ(1.0))
			s.SampleLight(c, mat, &s.Intersection, s.U, lightV, s.Distance)
		}
		s.FrameBuffer[screenIndex].AddPreMulColorSelf(c)
		s.ZBuffer[screenIndex] = s.Distance
	}
}
