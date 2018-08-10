package render

import (
	"github.com/tlyakhov/gofoom/core"
	"github.com/tlyakhov/gofoom/render/material"
	"github.com/tlyakhov/gofoom/render/state"
)

func WallMid(s *state.Slice) {
	mat := material.For(s.Segment.MidMaterial, s)

	for s.Y = s.ClippedStart; s.Y < s.ClippedEnd; s.Y++ {
		screenIndex := uint32(s.X + s.Y*s.ScreenWidth)

		if s.Distance >= s.ZBuffer[screenIndex] {
			continue
		}
		v := float64(s.Y-s.ScreenStart) / float64(s.ScreenEnd-s.ScreenStart)
		s.Intersection.Z = s.PhysicalSector.TopZ + v*(s.PhysicalSector.BottomZ-s.PhysicalSector.TopZ)

		light := s.Light(s.Intersection, s.Segment.Normal.To3D(), s.U, v)

		if s.Segment.MidBehavior == core.ScaleWidth || s.Segment.MidBehavior == core.ScaleNone {
			v = (v*(s.PhysicalSector.TopZ-s.PhysicalSector.BottomZ) - s.PhysicalSector.TopZ) / 64.0
		}

		u := s.U
		if s.Segment.MidBehavior == core.ScaleHeight || s.Segment.MidBehavior == core.ScaleNone {
			u = u * s.Segment.Length / 64.0
		}

		//fmt.Printf("%v\n", screenIndex)
		if mat != nil {
			s.Write(screenIndex, mat.Sample(u, v, light, s.ProjectZ(1.0)))
		}
		s.ZBuffer[screenIndex] = s.Distance
	}
}