package engine

import (
	"image/color"

	"github.com/tlyakhov/gofoom/util"
)

type Ray struct {
	Start, End *util.Vector2
}

type RenderSlice struct {
	*Renderer
	RenderTarget       []uint8
	X, Y, YStart, YEnd int
	TargetX            int
	Sector             *MapSector
	Segment            *MapSegment
	Ray                Ray
	RayIndex           int
	Intersection       *util.Vector3
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

func (s *RenderSlice) ProjectZ(z float64) float64 {
	return z * s.Renderer.viewFix[s.X] / s.Distance
}

func (s *RenderSlice) CalcScreen() {
	s.ProjHeightTop = s.ProjectZ(s.Sector.TopZ - s.CameraZ)
	s.ProjHeightBottom = s.ProjectZ(s.Sector.BottomZ - s.CameraZ)

	s.ScreenStart = s.ScreenHeight/2 - int(s.ProjHeightTop)
	s.ScreenEnd = s.ScreenHeight/2 - int(s.ProjHeightBottom)
	s.ClippedStart = util.Max(s.ScreenStart, s.YStart)
	s.ClippedEnd = util.Min(s.ScreenEnd, s.YEnd)
}

func (s *RenderSlice) Write(screenIndex uint, color color.NRGBA) {
	s.RenderTarget[screenIndex*4+0] = color.R
	s.RenderTarget[screenIndex*4+1] = color.G
	s.RenderTarget[screenIndex*4+2] = color.B
	s.RenderTarget[screenIndex*4+3] = color.A
}

func (s *RenderSlice) RenderMid() {
	if s.Segment.MidMaterial == nil {
		return
	}

	for s.Y = s.ClippedStart; s.Y < s.ClippedEnd; s.Y++ {
		screenIndex := uint(s.TargetX + s.Y*s.WorkerWidth)

		if s.Distance >= s.zbuffer[screenIndex] {
			continue
		}
		v := float64(s.Y-s.ScreenStart) / float64(s.ScreenEnd-s.ScreenStart)
		s.Intersection.Z = s.Sector.TopZ + v*(s.Sector.BottomZ-s.Sector.TopZ)

		// var light = this.map.light(slice.intersection, segment.normal, slice.sector, slice.segment, slice.u, v, true);

		if s.Segment.Midhavior == ScaleWidth || s.Segment.Midhavior == ScaleNone {
			v = (v*(s.Sector.TopZ-s.Sector.BottomZ) - s.Sector.TopZ) / 64.0
		}
		s.Write(screenIndex, s.Segment.MidMaterial.Sample(s, s.U, v, nil, uint(s.ScreenEnd-s.ScreenStart)))
		s.zbuffer[screenIndex] = s.Distance
	}
}
