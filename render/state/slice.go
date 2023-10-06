package state

import (
	"math"

	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/core"
	"tlyakhov/gofoom/entities"
)

type Ray struct {
	Start, End concepts.Vector2
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
	Intersection       concepts.Vector3
	Distance           float64
	LastPortalDistance float64
	U                  float64
	Depth              int
	CameraZ            float64
	FloorZ             float64
	CeilZ              float64
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
	s.FloorZ, s.CeilZ = s.PhysicalSector.CalcFloorCeilingZ(s.Intersection.To2D())
	s.ProjHeightTop = s.ProjectZ(s.CeilZ - s.CameraZ)
	s.ProjHeightBottom = s.ProjectZ(s.FloorZ - s.CameraZ)

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

func (s *Slice) Write(screenIndex uint32, c uint32) {
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

func (s *Slice) Light(world, normal concepts.Vector3, u, v float64) concepts.Vector3 {
	//return s.LightUnfiltered(world, normal, u, v)
	//le := LightElement{Sector: s.PhysicalSector, Segment: s.Segment, Normal: normal}
	//return le.Calculate(world, s.Segment)

	le00 := LightElement{Sector: s.PhysicalSector, Segment: s.Segment, Normal: normal}
	le10 := LightElement{Sector: s.PhysicalSector, Segment: s.Segment, Normal: normal}
	le11 := LightElement{Sector: s.PhysicalSector, Segment: s.Segment, Normal: normal}
	le01 := LightElement{Sector: s.PhysicalSector, Segment: s.Segment, Normal: normal}

	wall := s.Segment != nil && normal.Z == 0
	var lightmapLength uint32
	var wu, wv float64

	if !wall {
		if normal.Z < 0 {
			le00.Lightmap = s.PhysicalSector.CeilLightmap
			le10.Lightmap = s.PhysicalSector.CeilLightmap
			le11.Lightmap = s.PhysicalSector.CeilLightmap
			le01.Lightmap = s.PhysicalSector.CeilLightmap
		} else {
			le00.Lightmap = s.PhysicalSector.FloorLightmap
			le10.Lightmap = s.PhysicalSector.FloorLightmap
			le11.Lightmap = s.PhysicalSector.FloorLightmap
			le01.Lightmap = s.PhysicalSector.FloorLightmap
		}
		lightmapLength = uint32(len(le00.Lightmap))
		le00.MapIndex = s.PhysicalSector.LightmapAddress(world.To2D())
		le10.MapIndex = le00.MapIndex + 1
		le11.MapIndex = le10.MapIndex + s.PhysicalSector.LightmapWidth
		le01.MapIndex = le11.MapIndex - 1
		if le00.MapIndex > lightmapLength-1 {
			le00.MapIndex = lightmapLength - 1
		}
		if le10.MapIndex > lightmapLength-1 {
			le10.MapIndex = lightmapLength - 1
		}
		if le11.MapIndex > lightmapLength-1 {
			le11.MapIndex = lightmapLength - 1
		}
		if le01.MapIndex > lightmapLength-1 {
			le01.MapIndex = lightmapLength - 1
		}
		q := s.PhysicalSector.LightmapWorld(world, normal.Z > 0)
		wu = 1.0 - (world.X-q.X)/constants.LightGrid
		wv = 1.0 - (world.Y-q.Y)/constants.LightGrid
	} else {
		le00.Lightmap = s.Segment.Lightmap
		le10.Lightmap = s.Segment.Lightmap
		le11.Lightmap = s.Segment.Lightmap
		le01.Lightmap = s.Segment.Lightmap
		wu = u * (float64(s.Segment.LightmapWidth) - constants.LightSafety*2)
		wv = v * (float64(s.Segment.LightmapHeight) - constants.LightSafety*2)
		iu := uint32(wu) + constants.LightSafety
		iv := uint32(wv) + constants.LightSafety
		if iu > s.Segment.LightmapWidth-1 {
			iu = s.Segment.LightmapWidth - 1
		}
		if iv > s.Segment.LightmapHeight-1 {
			iv = s.Segment.LightmapHeight - 1
		}
		iu2 := iu + 1
		iv2 := iv + 1
		if iu2 > s.Segment.LightmapWidth-1 {
			iu2 = s.Segment.LightmapWidth - 1
		}
		if iv2 > s.Segment.LightmapHeight-1 {
			iv2 = s.Segment.LightmapHeight - 1
		}
		le00.MapIndex = iu + iv*s.Segment.LightmapWidth
		le10.MapIndex = iu2 + iv*s.Segment.LightmapWidth
		le11.MapIndex = iu2 + iv2*s.Segment.LightmapWidth
		le01.MapIndex = iu + iv2*s.Segment.LightmapWidth
		wu = 1.0 - (wu - math.Floor(wu))
		wv = 1.0 - (wv - math.Floor(wv))
	}
	//wu = concepts.Clamp(wu, 0.0, 1.0)
	//wv = concepts.Clamp(wv, 0.0, 1.0)

	return le00.Get(wall).Mul(wu * wv).
		Add(le10.Get(wall).Mul((1.0 - wu) * wv)).
		Add(le11.Get(wall).Mul((1.0 - wu) * (1.0 - wv))).
		Add(le01.Get(wall).Mul(wu * (1.0 - wv)))
}

func (s *Slice) LightUnfiltered(world, normal concepts.Vector3, u, v float64) concepts.Vector3 {
	le00 := LightElement{Sector: s.PhysicalSector, Segment: s.Segment, Normal: normal}
	wall := s.Segment != nil && normal.Z == 0

	if !wall {
		if normal.Z < 0 {
			le00.Lightmap = s.PhysicalSector.CeilLightmap
		} else {
			le00.Lightmap = s.PhysicalSector.FloorLightmap
		}
		lightmapLength := uint32(len(le00.Lightmap))
		le00.MapIndex = s.PhysicalSector.LightmapAddress(world.To2D())
		if le00.MapIndex > lightmapLength-1 {
			le00.MapIndex = lightmapLength - 1
		}
	} else {
		le00.Lightmap = s.Segment.Lightmap
		lightmapLength := uint32(len(le00.Lightmap))
		le00.MapIndex = s.Segment.LightmapAddress(u, v)
		if le00.MapIndex > lightmapLength-1 {
			le00.MapIndex = lightmapLength - 1
		}
	}

	return le00.Get(wall)
}
