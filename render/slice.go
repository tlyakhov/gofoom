package render

import (
	"image/color"
	"math"

	"github.com/tlyakhov/gofoom/mapping"
	"github.com/tlyakhov/gofoom/math"
)

type Ray struct {
	Start, End *math.Vector2
}

type Slice struct {
	*Config
	RenderTarget       []uint8
	X, Y, YStart, YEnd int
	TargetX            int
	Sector             *mapping.MapSector
	Segment            *mapping.MapSegment
	Ray                Ray
	RayIndex           int
	Intersection       *math.Vector3
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
	s.ClippedStart = math.Max(s.ScreenStart, s.YStart)
	s.ClippedEnd = math.Min(s.ScreenEnd, s.YEnd)
}

func (s *Slice) Write(screenIndex uint, color color.NRGBA) {
	s.RenderTarget[screenIndex*4+0] = color.R
	s.RenderTarget[screenIndex*4+1] = color.G
	s.RenderTarget[screenIndex*4+2] = color.B
	s.RenderTarget[screenIndex*4+3] = color.A
}

func (s *Slice) RenderMid() {
	if s.Segment.MidMaterial == nil {
		return
	}

	for s.Y = s.ClippedStart; s.Y < s.ClippedEnd; s.Y++ {
		screenIndex := uint(s.TargetX + s.Y*s.WorkerWidth)

		if s.Distance >= s.ZBuffer[screenIndex] {
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

func (s *Slice) RenderFloor() {
	world := &math.Vector3{0, 0, s.Sector.BottomZ}

	for s.Y = s.ClippedEnd; s.Y < s.YEnd; s.Y++ {
		if s.Y-s.ScreenHeight/2 == 0 {
			continue
		}

		distToFloor := (-s.Sector.BottomZ + s.CameraZ) * s.ViewFix[s.X] / float64(s.Y-s.ScreenHeight/2)
		scaler := float64(s.Sector.FloorMaterial.Texture.Height) * s.Sector.FloorScale / distToFloor
		screenIndex := uint(s.TargetX + s.Y*s.WorkerWidth)

		if distToFloor >= s.zbuffer[screenIndex] {
			continue
		}

		world.X = s.Map.Player.Pos.X + s.TrigTable[s.RayIndex].cos*distToFloor
		world.Y = s.Map.Player.Pos.Y + s.TrigTable[s.RayIndex].sin*distToFloor

		tx := world.X / s.Sector.FloorScale
		tx -= math.Floor(tx)
		ty := world.Y / s.Sector.FloorScale
		ty -= math.Floor(ty)
		if tx < 0 {
			tx += 1.0
		}
		if ty < 0 {
			ty += 1.0
		}

		// var light = this.map.light(world, FLOOR_NORMAL, slice.sector, slice.segment, null, null, true);

		s.Write(screenIndex, s.Sector.FloorMaterial.Sample(s, tx, ty, nil, uint(scaler)))
		s.ZBuffer[screenIndex] = distToFloor
	}
}

func (s *Slice) RenderCeiling() {
	world := &math.Vector3{0, 0, s.Sector.TopZ}

	for s.Y = s.YStart; s.Y < s.ClippedStart; s.Y++ {
		if s.Y-s.ScreenHeight/2 == 0 {
			continue
		}

		distToCeil := (s.Sector.TopZ - s.CameraZ) * s.ViewFix[s.X] / float64(s.ScreenHeight/2-1-s.Y)
		scaler := float64(s.Sector.CeilMaterial.Texture.Height) * s.Sector.CeilScale / distToCeil
		screenIndex := uint(s.TargetX + s.Y*s.WorkerWidth)

		if distToCeil >= s.zbuffer[screenIndex] {
			continue
		}

		world.X = s.Map.Player.Pos.X + s.TrigTable[s.RayIndex].cos*distToCeil
		world.Y = s.Map.Player.Pos.Y + s.TrigTable[s.RayIndex].sin*distToCeil

		tx := world.X / s.Sector.CeilScale
		tx -= math.Floor(tx)
		ty := world.Y / s.Sector.CeilScale
		ty -= math.Floor(ty)
		tx = math.Abs(tx)
		ty = math.Abs(ty)
		// var light = this.map.light(world, CEIL_NORMAL, slice.sector, slice.segment, null, null, true);

		s.Write(screenIndex, s.Sector.CeilMaterial.Sample(s, tx, ty, nil, uint(scaler)))
		s.ZBuffer[screenIndex] = distToCeil
	}
}
