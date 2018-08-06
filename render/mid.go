package render

import (
	"github.com/tlyakhov/gofoom/mapping"
	"github.com/tlyakhov/gofoom/render/material"
	"github.com/tlyakhov/gofoom/render/state"
)

func WallMid(s *state.Slice) {
	mat := material.For(s.Segment.MidMaterial, s)

	for s.Y = s.ClippedStart; s.Y < s.ClippedEnd; s.Y++ {
		screenIndex := uint(s.TargetX + s.Y*s.WorkerWidth)

		if s.Distance >= s.ZBuffer[screenIndex] {
			continue
		}
		v := float64(s.Y-s.ScreenStart) / float64(s.ScreenEnd-s.ScreenStart)
		s.Intersection.Z = s.Sector.TopZ + v*(s.Sector.BottomZ-s.Sector.TopZ)

		// var light = this.map.light(slice.intersection, segment.normal, slice.sector, slice.segment, slice.u, v, true);

		if s.Segment.MidBehavior == mapping.ScaleWidth || s.Segment.MidBehavior == mapping.ScaleNone {
			v = (v*(s.Sector.TopZ-s.Sector.BottomZ) - s.Sector.TopZ) / 64.0
		}

		u := s.U
		if s.Segment.MidBehavior == mapping.ScaleHeight || s.Segment.MidBehavior == mapping.ScaleNone {
			u = u * s.Segment.Length / 64.0
		}

		//fmt.Printf("%v\n", screenIndex)
		if mat != nil {
			s.Write(screenIndex, mat.Sample(u, v, nil, s.ProjectZ(1.0)))
		}
		s.ZBuffer[screenIndex] = s.Distance
	}
}
