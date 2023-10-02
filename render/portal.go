package render

import (
	"tlyakhov/gofoom/core"
	"tlyakhov/gofoom/render/material"
	"tlyakhov/gofoom/render/state"
)

// WallHi renders the top portion of a portal segment.
func WallHi(s *state.SlicePortal) {
	mat := material.For(s.AdjSegment.HiMaterial, s.Slice)

	u := s.U
	if s.Segment.HiBehavior == core.ScaleHeight || s.Segment.HiBehavior == core.ScaleNone {
		if s.PhysicalSector.Winding < 0 {
			u = 1.0 - u
		}
		u = (s.Segment.P.X + s.Segment.P.Y + u*s.Segment.Length) / 64.0
	}

	for s.Y = s.ClippedStart; s.Y < s.AdjClippedTop; s.Y++ {
		screenIndex := uint32(s.X + s.Y*s.ScreenWidth)
		if s.Distance >= s.ZBuffer[screenIndex] {
			continue
		}
		v := float64(s.Y-s.ScreenStart) / float64(s.AdjScreenTop-s.ScreenStart)
		s.Intersection.Z = s.CeilZ - v*(s.CeilZ-s.AdjCeilZ)

		light := s.Light(s.Intersection, s.Segment.Normal.To3D(), s.U, v*0.5)

		if s.Segment.HiBehavior == core.ScaleWidth || s.Segment.HiBehavior == core.ScaleNone {
			v = (s.AdjCeilZ - v*(s.AdjCeilZ-s.CeilZ)) / 64.0
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
	u := s.U
	if s.Segment.LoBehavior == core.ScaleHeight || s.Segment.LoBehavior == core.ScaleNone {
		if s.PhysicalSector.Winding < 0 {
			u = 1.0 - u
		}
		u = (s.Segment.P.X + s.Segment.P.Y + u*s.Segment.Length) / 64.0
	}
	for s.Y = s.AdjClippedBottom; s.Y < s.ClippedEnd; s.Y++ {
		screenIndex := uint32(s.X + s.Y*s.ScreenWidth)
		if s.Distance >= s.ZBuffer[screenIndex] {
			continue
		}
		v := float64(s.Y-s.AdjClippedBottom) / float64(s.ScreenEnd-s.AdjScreenBottom)
		s.Intersection.Z = s.AdjFloorZ - v*(s.AdjFloorZ-s.FloorZ)
		light := s.Light(s.Intersection, s.Segment.Normal.To3D(), s.U, v*0.5+0.5)

		if s.Segment.LoBehavior == core.ScaleWidth || s.Segment.LoBehavior == core.ScaleNone {
			v = (v*(s.FloorZ-s.AdjFloorZ) - s.FloorZ) / 64.0
		}

		if mat != nil {
			s.Write(screenIndex, mat.Sample(u, v, light, s.ProjectZ(1.0)))
		}
		s.ZBuffer[screenIndex] = s.Distance
	}
}
