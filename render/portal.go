package render

import (
	"github.com/tlyakhov/gofoom/core"
	"github.com/tlyakhov/gofoom/render/material"
	"github.com/tlyakhov/gofoom/render/state"
)

// WallHi renders the top portion of a portal segment.
func WallHi(s *state.SlicePortal) {
	mat := material.For(s.AdjSegment.HiMaterial, s.Slice)

	for s.Y = s.ClippedStart; s.Y < s.AdjClippedTop; s.Y++ {
		screenIndex := uint32(s.X + s.Y*s.ScreenWidth)
		if s.Distance >= s.ZBuffer[screenIndex] {
			continue
		}
		v := float64(s.Y-s.ScreenStart) / float64(s.AdjScreenTop-s.ScreenStart)
		s.Intersection.Z = s.PhysicalSector.TopZ - v*(s.PhysicalSector.TopZ-s.Adj.Physical().TopZ)

		light := s.Light(s.Intersection, s.Segment.Normal.To3D(), s.U, v*0.5)

		if s.Segment.HiBehavior == core.ScaleHeight || s.Segment.HiBehavior == core.ScaleAll {
			v = (s.Adj.Physical().TopZ - v*(s.Adj.Physical().TopZ-s.PhysicalSector.TopZ)) / 64.0
		}

		u := s.U
		if s.Segment.HiBehavior == core.ScaleHeight || s.Segment.HiBehavior == core.ScaleNone {
			u = u * s.Segment.Length / 64.0
		}

		if mat != nil {
			s.Write(screenIndex, mat.Sample(u, v, light, s.ProjectZ(1.0)))
		}
		s.ZBuffer[screenIndex] = s.Distance
	}
}

// WallLow renders the bottom portion of a portal segment.
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

		if s.Segment.LoBehavior == core.ScaleHeight || s.Segment.LoBehavior == core.ScaleAll {
			v = (v*(s.PhysicalSector.BottomZ-s.Adj.Physical().BottomZ) - s.PhysicalSector.BottomZ) / 64.0
		}

		u := s.U
		if s.Segment.LoBehavior == core.ScaleHeight || s.Segment.LoBehavior == core.ScaleNone {
			u = u * s.Segment.Length / 64.0
		}

		if mat != nil {
			s.Write(screenIndex, mat.Sample(u, v, light, s.ProjectZ(1.0)))
		}
		s.ZBuffer[screenIndex] = s.Distance
	}
}
