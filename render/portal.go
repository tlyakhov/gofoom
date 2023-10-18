package render

import (
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/core"
	"tlyakhov/gofoom/render/state"
)

func WallHiPick(s *state.SlicePortal) {
	if s.Y >= s.ClippedStart && s.Y < s.AdjClippedTop {
		s.PickedElements = append(s.PickedElements, state.PickedElement{Type: "hi", ISerializable: s.Segment})
	}
}

// WallHi renders the top portion of a portal segment.
func WallHi(s *state.SlicePortal) {
	mat := s.AdjSegment.HiMaterial
	u := s.U
	if s.Segment.HiBehavior == core.ScaleHeight || s.Segment.HiBehavior == core.ScaleNone {
		if s.PhysicalSector.Winding < 0 {
			u = 1.0 - u
		}
		u = (s.Segment.P[0] + s.Segment.P[1] + u*s.Segment.Length) / 64.0
	}
	light := &concepts.Vector3{}
	for s.Y = s.ClippedStart; s.Y < s.AdjClippedTop; s.Y++ {
		screenIndex := uint32(s.X + s.Y*s.ScreenWidth)
		if s.Distance >= s.ZBuffer[screenIndex] {
			continue
		}
		v := float64(s.Y-s.ScreenStart) / float64(s.AdjScreenTop-s.ScreenStart)
		s.Intersection[2] = (1.0-v)*s.CeilZ + v*s.AdjCeilZ
		lightV := (s.PhysicalSector.Max[2] - s.Intersection[2]) / (s.PhysicalSector.Max[2] - s.PhysicalSector.Min[2])
		s.Light(light, &s.Intersection, s.U, lightV, s.Distance)

		if s.Segment.HiBehavior == core.ScaleWidth || s.Segment.HiBehavior == core.ScaleNone {
			v = s.Intersection[2] / 64.0
		}

		if mat != nil {
			s.Write(screenIndex, s.SampleMaterial(mat, u, v, light, s.ProjectZ(1.0)))
		}
		s.ZBuffer[screenIndex] = s.Distance
	}
}

func WallLowPick(s *state.SlicePortal) {
	if s.Y >= s.AdjClippedBottom && s.Y < s.ClippedEnd {
		s.PickedElements = append(s.PickedElements, state.PickedElement{Type: "lo", ISerializable: s.Segment})
	}
}

// WallLow renders the bottom portion of a portal segment.
func WallLow(s *state.SlicePortal) {
	mat := s.AdjSegment.LoMaterial
	u := s.U
	if s.Segment.LoBehavior == core.ScaleHeight || s.Segment.LoBehavior == core.ScaleNone {
		if s.PhysicalSector.Winding < 0 {
			u = 1.0 - u
		}
		u = (s.Segment.P[0] + s.Segment.P[1] + u*s.Segment.Length) / 64.0
	}
	light := concepts.Vector3{}
	for s.Y = s.AdjClippedBottom; s.Y < s.ClippedEnd; s.Y++ {
		screenIndex := uint32(s.X + s.Y*s.ScreenWidth)
		if s.Distance >= s.ZBuffer[screenIndex] {
			continue
		}
		v := float64(s.Y-s.AdjScreenBottom) / float64(s.ScreenEnd-s.AdjScreenBottom)
		s.Intersection[2] = (1.0-v)*s.AdjFloorZ + v*s.FloorZ
		lightV := (s.PhysicalSector.Max[2] - s.Intersection[2]) / (s.PhysicalSector.Max[2] - s.PhysicalSector.Min[2])
		s.Light(&light, &s.Intersection, s.U, lightV, s.Distance)

		if s.Segment.LoBehavior == core.ScaleWidth || s.Segment.LoBehavior == core.ScaleNone {
			v = s.Intersection[2] / 64.0
		}

		if mat != nil {
			s.Write(screenIndex, s.SampleMaterial(mat, u, v, &light, s.ProjectZ(1.0)))
		}
		s.ZBuffer[screenIndex] = s.Distance
	}
}
