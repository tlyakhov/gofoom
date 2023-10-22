package state

import (
	"math"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/core"
	"tlyakhov/gofoom/materials"
	"tlyakhov/gofoom/mobs"
)

type Ray struct {
	Start, End concepts.Vector2
}

type PickedElement struct {
	Type string // ceil, floor, mid, hi, lo, mob
	concepts.ISerializable
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
	Pick               bool
	PickedElements     []PickedElement
}

func (s *Slice) ProjectZ(z float64) float64 {
	return z * s.ViewFix[s.X] / s.Distance
}

func (s *Slice) CalcScreen() {
	// Screen slice precalculation
	s.FloorZ, s.CeilZ = s.PhysicalSector.SlopedZRender(s.Intersection.To2D())
	s.ProjHeightTop = s.ProjectZ(s.CeilZ - s.CameraZ)
	s.ProjHeightBottom = s.ProjectZ(s.FloorZ - s.CameraZ)

	s.ScreenStart = s.ScreenHeight/2 - int(s.ProjHeightTop)
	s.ScreenEnd = s.ScreenHeight/2 - int(s.ProjHeightBottom)
	s.ClippedStart = concepts.IntClamp(s.ScreenStart, s.YStart, s.YEnd)
	s.ClippedEnd = concepts.IntClamp(s.ScreenEnd, s.YStart, s.YEnd)
	// Frame Tint precalculation
	tint := s.Map.Player.(*mobs.Player).FrameTint
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

func (s *Slice) SampleMaterial(m core.Sampleable, u, v float64, light *concepts.Vector3, scale float64) uint32 {
	if sampled, ok := m.(*materials.Sampled); ok {
		if sampled.IsLiquid {
			u += math.Cos(float64(s.Frame)*constants.LiquidChurnSpeed*concepts.Deg2rad) * constants.LiquidChurnSize
			v += math.Sin(float64(s.Frame)*constants.LiquidChurnSpeed*concepts.Deg2rad) * constants.LiquidChurnSize
		}
	}

	if sky, ok := m.(*materials.Sky); ok {
		v = float64(s.Y) / (float64(s.ScreenHeight) - 1)

		if sky.StaticBackground {
			u = float64(s.X) / (float64(s.ScreenWidth) - 1)
		} else {
			u = s.Angle / (2.0 * math.Pi)
			for ; u < 0; u++ {
			}
			for ; u > 1; u-- {
			}
		}
	}

	return m.Sample(u, v, light, scale)

}
func (s *Slice) Light(result, world *concepts.Vector3, u, v, dist float64) *concepts.Vector3 {
	/*// testing...
	result[0] = concepts.Clamp(dist/500.0, 0.0, 1.0)
	result[1] = result[0]
	result[2] = result[0]
	return result*/

	// Don't filter far away lightmaps. Tolerate a ~5px snap-in
	if dist > float64(s.ScreenWidth)*constants.LightGrid*0.2 {
		return s.LightUnfiltered(result, world, u, v)
	}

	le00 := &s.LightElements[0]
	le10 := &s.LightElements[1]
	le11 := &s.LightElements[2]
	le01 := &s.LightElements[3]

	wall := s.Segment != nil && s.Normal[2] == 0
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

	r00 := le00.Get(wall)
	r10 := le10.Get(wall)
	r11 := le11.Get(wall)
	r01 := le01.Get(wall)
	result[0] = r00[0]*(wu*wv) + r10[0]*((1.0-wu)*wv) + r11[0]*(1.0-wu)*(1.0-wv) + r01[0]*wu*(1.0-wv)
	result[1] = r00[1]*(wu*wv) + r10[1]*((1.0-wu)*wv) + r11[1]*(1.0-wu)*(1.0-wv) + r01[1]*wu*(1.0-wv)
	result[2] = r00[2]*(wu*wv) + r10[2]*((1.0-wu)*wv) + r11[2]*(1.0-wu)*(1.0-wv) + r01[2]*wu*(1.0-wv)

	return result
}

func (s *Slice) LightUnfiltered(result, world *concepts.Vector3, u, v float64) *concepts.Vector3 {
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
		le.MapIndex = s.PhysicalSector.LightmapAddress(world.To2D())
	} else {
		s.Lightmap = s.Segment.Lightmap
		s.LightmapAge = s.Segment.LightmapAge
		le.MapIndex = s.Segment.LightmapAddress(u, v)
	}

	r00 := le.Get(wall)
	result[0] = r00[0]
	result[1] = r00[1]
	result[2] = r00[2]
	return result
}
