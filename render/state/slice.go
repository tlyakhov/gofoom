package state

import (
	"image/color"

	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/mapping"
)

type Ray struct {
	Start, End *concepts.Vector2
}

type Slice struct {
	*Config
	RenderTarget       []uint8
	X, Y, YStart, YEnd int
	TargetX            int
	Map                *mapping.Map
	Sector             *mapping.Sector
	Segment            *mapping.Segment
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
}

func (s *Slice) ProjectZ(z float64) float64 {
	return z * s.ViewFix[s.X] / s.Distance
}

func (s *Slice) CalcScreen() {
	s.ProjHeightTop = s.ProjectZ(s.Sector.TopZ - s.CameraZ)
	s.ProjHeightBottom = s.ProjectZ(s.Sector.BottomZ - s.CameraZ)

	s.ScreenStart = s.ScreenHeight/2 - int(s.ProjHeightTop)
	s.ScreenEnd = s.ScreenHeight/2 - int(s.ProjHeightBottom)
	s.ClippedStart = concepts.IntClamp(s.ScreenStart, s.YStart, s.YEnd)
	s.ClippedEnd = concepts.IntClamp(s.ScreenEnd, s.YStart, s.YEnd)
}

func (s *Slice) Write(screenIndex uint, color color.NRGBA) {
	s.RenderTarget[screenIndex*4+0] = color.R
	s.RenderTarget[screenIndex*4+1] = color.G
	s.RenderTarget[screenIndex*4+2] = color.B
	s.RenderTarget[screenIndex*4+3] = color.A
}
