package render

import (
	"github.com/tlyakhov/gofoom/core"
	"github.com/tlyakhov/gofoom/render/material"
	"github.com/tlyakhov/gofoom/render/state"
)

// WallMid renders the wall portion of a non-portal segment.
func WallMid(s *state.Slice) {
	mat := material.For(s.Segment.MidMaterial, s)

	u := s.U
	if s.Segment.MidBehavior == core.ScaleHeight || s.Segment.MidBehavior == core.ScaleNone {
		if s.PhysicalSector.Winding < 0 {
			u = 1.0 - u
		}
		u = (s.Segment.P.X + s.Segment.P.Y + u*s.Segment.Length) / 64.0
	}

	for s.Y = s.ClippedStart; s.Y < s.ClippedEnd; s.Y++ {
		screenIndex := uint32(s.X + s.Y*s.ScreenWidth)

		if s.Distance >= s.ZBuffer[screenIndex] {
			continue
		}
		v := float64(s.Y-s.ScreenStart) / float64(s.ScreenEnd-s.ScreenStart)
		s.Intersection.Z = s.CeilZ + v*(s.FloorZ-s.CeilZ)

		light := s.Light(s.Intersection, s.Segment.Normal.To3D(), s.U, v)

		if s.Segment.MidBehavior == core.ScaleWidth || s.Segment.MidBehavior == core.ScaleNone {
			v = (v*(s.CeilZ-s.FloorZ) - s.CeilZ) / 64.0
		}

		//fmt.Printf("%v\n", screenIndex)
		if mat != nil {
			s.Write(screenIndex, mat.Sample(u, v, light, s.ProjectZ(1.0)))
		}
		s.ZBuffer[screenIndex] = s.Distance
	}
}
