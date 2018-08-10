package render

import (
	"github.com/tlyakhov/gofoom/core"
	"github.com/tlyakhov/gofoom/render/material"
	"github.com/tlyakhov/gofoom/render/state"
)

func WallHi(s *state.SlicePortal) {
	mat := material.For(s.AdjSegment.HiMaterial, s.Slice)

	for s.Y = s.ClippedStart; s.Y < s.AdjClippedTop; s.Y++ {
		screenIndex := uint32(s.X + s.Y*s.ScreenWidth)
		//fmt.Printf("%v %v %v %v\n", screenIndex, s.Y, s.ClippedStart, s.AdjClippedTop)
		if s.Distance >= s.ZBuffer[screenIndex] {
			continue
		}
		v := float64(s.Y-s.ScreenStart) / float64(s.AdjScreenTop-s.ScreenStart)
		s.Intersection.Z = s.PhysicalSector.TopZ - v*(s.PhysicalSector.TopZ-s.Adj.Physical().TopZ)

		// var light = this.map.light(s.intersection, segment.normal, s.sector, s.segment, s.u, v * 0.5, true);
		light := s.Light(s.Intersection, s.Segment.Normal.To3D(), s.U, v*0.5)

		if s.AdjSegment.HiBehavior == core.ScaleWidth || s.AdjSegment.HiBehavior == core.ScaleNone {
			v = (s.Adj.Physical().TopZ - v*(s.Adj.Physical().TopZ-s.PhysicalSector.TopZ)) / 64.0
			//fmt.Printf("%v\n", v)
		}
		if mat != nil {
			s.Write(screenIndex, mat.Sample(s.U, v, light, s.ProjectZ(1.0)))
		}
		s.ZBuffer[screenIndex] = s.Distance
	}
}

func WallLow(s *state.SlicePortal) {
	mat := material.For(s.AdjSegment.LoMaterial, s.Slice)

	for s.Y = s.AdjClippedBottom; s.Y < s.ClippedEnd; s.Y++ {
		screenIndex := uint32(s.X + s.Y*s.ScreenWidth)
		if s.Distance >= s.ZBuffer[screenIndex] {
			continue
		}
		v := float64(s.Y-s.AdjClippedBottom) / float64(s.ScreenEnd-s.AdjScreenBottom)
		s.Intersection.Z = s.Adj.Physical().BottomZ - v*(s.Adj.Physical().BottomZ-s.PhysicalSector.BottomZ)
		light := s.Light(s.Intersection, s.Segment.Normal.To3D(), s.U, v*0.5+0.5)
		if s.AdjSegment.LoBehavior == core.ScaleWidth || s.AdjSegment.LoBehavior == core.ScaleNone {
			v = (v*(s.PhysicalSector.BottomZ-s.Adj.Physical().BottomZ) - s.PhysicalSector.BottomZ) / 64.0
		}

		if mat != nil {
			s.Write(screenIndex, mat.Sample(s.U, v, light, s.ProjectZ(1.0)))
		}
		s.ZBuffer[screenIndex] = s.Distance
	}
}
