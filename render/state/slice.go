package state

import (
	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/core"
	"github.com/tlyakhov/gofoom/entities"
)

type Ray struct {
	Start, End *concepts.Vector2
}

type Slice struct {
	*Config
	RenderTarget       []uint8
	X, Y, YStart, YEnd int
	Map                *core.Map
	PhysicalSector     *core.PhysicalSector
	Segment            *core.Segment
	Ray                *Ray
	Angle              float64
	AngleCos           float64
	AngleSin           float64
	Intersection       *concepts.Vector3
	Distance           float64
	U                  float64
	Depth              int
	CameraZ            float64
	ProjHeightTop      float64
	ProjHeightBottom   float64
	ScreenStart        int
	ScreenEnd          int
	ClippedStart       int
	ClippedEnd         int
	FrameTint          [4]uint32
}

func (s *Slice) ProjectZ(z float64) float64 {
	return z * s.ViewFix[s.X] / s.Distance
}

func (s *Slice) CalcScreen() {
	// Screen slice precalculation
	s.ProjHeightTop = s.ProjectZ(s.PhysicalSector.TopZ - s.CameraZ)
	s.ProjHeightBottom = s.ProjectZ(s.PhysicalSector.BottomZ - s.CameraZ)

	s.ScreenStart = s.ScreenHeight/2 - int(s.ProjHeightTop)
	s.ScreenEnd = s.ScreenHeight/2 - int(s.ProjHeightBottom)
	s.ClippedStart = concepts.IntClamp(s.ScreenStart, s.YStart, s.YEnd)
	s.ClippedEnd = concepts.IntClamp(s.ScreenEnd, s.YStart, s.YEnd)
	// Frame Tint precalculation
	tint := s.Map.Player.(*entities.Player).FrameTint
	s.FrameTint[0] = uint32(tint.R) * uint32(tint.A)
	s.FrameTint[1] = uint32(tint.G) * uint32(tint.A)
	s.FrameTint[2] = uint32(tint.B) * uint32(tint.A)
	s.FrameTint[3] = uint32(0xFF - tint.A)
}

func (s *Slice) Write(screenIndex uint, c uint32) {
	if s.FrameTint[3] != 0xFF {
		s.RenderTarget[screenIndex*4+0] = uint8((((c>>24)&0xFF)*s.FrameTint[3] + s.FrameTint[0]) >> 8)
		s.RenderTarget[screenIndex*4+1] = uint8((((c>>16)&0xFF)*s.FrameTint[3] + s.FrameTint[1]) >> 8)
		s.RenderTarget[screenIndex*4+2] = uint8((((c>>8)&0xFF)*s.FrameTint[3] + s.FrameTint[2]) >> 8)
	} else {
		s.RenderTarget[screenIndex*4+0] = uint8((c >> 24) & 0xFF)
		s.RenderTarget[screenIndex*4+1] = uint8((c >> 16) & 0xFF)
		s.RenderTarget[screenIndex*4+2] = uint8((c >> 8) & 0xFF)
	}
	s.RenderTarget[screenIndex*4+3] = uint8(c & 0xFF)
}
