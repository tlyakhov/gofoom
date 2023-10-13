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
	LightElements      [4]LightElement
	Normal             concepts.Vector3
	Lightmap           []concepts.Vector3
	LightmapAge        []int
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

func (s *Slice) Light(world *concepts.Vector3, u, v float64) *concepts.Vector3 {
	//return s.LightUnfiltered(world, u, v)
	//le := LightElement{Sector: s.PhysicalSector, Segment: s.Segment, Normal: normal}
	//return le.Calculate(world, s.Segment)

	le00 := &s.LightElements[0]
	le10 := &s.LightElements[1]
	le11 := &s.LightElements[2]
	le01 := &s.LightElements[3]

	wall := s.Segment != nil && s.Normal[2] == 0
	var lightmapLength uint32
	var wu, wv float64

	if !wall {
		if s.Normal[2] < 0 {
			s.Lightmap = s.PhysicalSector.CeilLightmap
			s.LightmapAge = s.PhysicalSector.CeilLightmapAge
		} else {
			s.Lightmap = s.PhysicalSector.FloorLightmap
			s.LightmapAge = s.PhysicalSector.FloorLightmapAge
		}
		le00.MapIndex = s.PhysicalSector.LightmapAddress(world.To2D())
		le10.MapIndex = le00.MapIndex + 1
		le11.MapIndex = le10.MapIndex + s.PhysicalSector.LightmapWidth
		le01.MapIndex = le11.MapIndex - 1
		q := &concepts.Vector3{world[0], world[1], world[2]}
		s.PhysicalSector.ToLightmapWorld(q, s.Normal[2] > 0)
		wu = 1.0 - (world[0]-q[0])/constants.LightGrid
		wv = 1.0 - (world[1]-q[1])/constants.LightGrid
	} else {
		s.Lightmap = s.Segment.Lightmap
		s.LightmapAge = s.Segment.LightmapAge
		le00.MapIndex = s.Segment.LightmapAddress(u, v)
		le10.MapIndex = le00.MapIndex + 1
		le11.MapIndex = le10.MapIndex + s.Segment.LightmapWidth
		le01.MapIndex = le11.MapIndex - 1
		wu = u * float64(s.Segment.LightmapWidth-constants.LightSafety*2)
		wv = v * float64(s.Segment.LightmapHeight-constants.LightSafety*2)
		wu = 1.0 - (wu - math.Floor(wu))
		wv = 1.0 - (wv - math.Floor(wv))
	}

	lightmapLength = uint32(len(s.Lightmap))
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
	r00 := *le00.Get(wall)
	r10 := *le10.Get(wall)
	r11 := *le11.Get(wall)
	r01 := *le01.Get(wall)
	return r00.MulSelf(wu * wv).
		AddSelf(r10.MulSelf((1.0 - wu) * wv)).
		AddSelf(r11.MulSelf((1.0 - wu) * (1.0 - wv))).
		AddSelf(r01.MulSelf(wu * (1.0 - wv)))
}

func (s *Slice) LightUnfiltered(world *concepts.Vector3, u, v float64) *concepts.Vector3 {
	le := &s.LightElements[0]
	wall := s.Segment != nil && s.Normal[2] == 0

	if !wall {
		if s.Normal[2] < 0 {
			s.Lightmap = s.PhysicalSector.CeilLightmap
			s.LightmapAge = s.PhysicalSector.CeilLightmapAge
		} else {
			s.Lightmap = s.PhysicalSector.FloorLightmap
			s.LightmapAge = s.PhysicalSector.FloorLightmapAge
		}
		lightmapLength := uint32(len(le.Lightmap))
		le.MapIndex = s.PhysicalSector.LightmapAddress(world.To2D())
		if le.MapIndex > lightmapLength-1 {
			le.MapIndex = lightmapLength - 1
		}
	} else {
		s.Lightmap = s.Segment.Lightmap
		s.LightmapAge = s.Segment.LightmapAge
		lightmapLength := uint32(len(le.Lightmap))
		le.MapIndex = s.Segment.LightmapAddress(u, v)
		if le.MapIndex > lightmapLength-1 {
			le.MapIndex = lightmapLength - 1
		}
	}

	return le.Get(wall)
}
