package state

import (
	"math"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
)

type Ray struct {
	Start, End, Delta concepts.Vector2
}

type PickedType int

//go:generate go run github.com/dmarkham/enumer -type=PickedType -json
const (
	PickCeiling PickedType = iota
	PickHigh
	PickMid
	PickLow
	PickFloor
	PickBody
)

type PickedElement struct {
	Type    PickedType // ceil, floor, mid, hi, lo, body
	Element any
}

type Column struct {
	*Config
	X, Y, YStart, YEnd int
	Sector             *core.Sector
	Segment            *core.SectorSegment
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
	LightElements      [4]LightElement
	Normal             concepts.Vector3
	MaterialColor      concepts.Vector4
	Light              concepts.Vector4
	Lightmap           []concepts.Vector3
	LightmapAge        []int
	Pick               bool
	PickedElements     []PickedElement
}

func (c *Column) ProjectZ(z float64) float64 {
	return z * c.ViewFix[c.X] / c.Distance
}

func (c *Column) CalcScreen() {
	// Screen slice precalculation
	c.FloorZ, c.CeilZ = c.Sector.SlopedZRender(c.Intersection.To2D())
	c.ProjHeightTop = c.ProjectZ(c.CeilZ - c.CameraZ)
	c.ProjHeightBottom = c.ProjectZ(c.FloorZ - c.CameraZ)

	c.ScreenStart = c.ScreenHeight/2 - int(c.ProjHeightTop)
	c.ScreenEnd = c.ScreenHeight/2 - int(c.ProjHeightBottom)
	c.ClippedStart = concepts.IntClamp(c.ScreenStart, c.YStart, c.YEnd)
	c.ClippedEnd = concepts.IntClamp(c.ScreenEnd, c.YStart, c.YEnd)
}

func (c *Column) SampleShader(ishader *concepts.EntityRef, extraStages []*materials.ShaderStage, u, v float64, scale float64) *concepts.Vector4 {
	c.MaterialColor[0] = 0
	c.MaterialColor[1] = 0
	c.MaterialColor[2] = 0
	c.MaterialColor[3] = 0
	shader := materials.ShaderFromDb(ishader)
	if shader == nil {
		c.sampleTexture(&c.MaterialColor, ishader, nil, u, v, scale)
	} else {
		for _, stage := range shader.Stages {
			c.sampleTexture(&c.MaterialColor, stage.Texture, stage, u, v, scale)
		}
	}

	for _, stage := range extraStages {
		c.sampleTexture(&c.MaterialColor, stage.Texture, stage, u, v, scale)
	}
	return &c.MaterialColor
}

func (c *Column) sampleTexture(result *concepts.Vector4, material *concepts.EntityRef, stage *materials.ShaderStage, u, v float64, scale float64) *concepts.Vector4 {
	// Should refactor this scale thing, it's hard to reason about
	scaleDivisor := 1.0

	if stage != nil {
		u, v = stage.Transform[0]*u+stage.Transform[2]*v+stage.Transform[4], stage.Transform[1]*u+stage.Transform[3]*v+stage.Transform[5]
		if (stage.Flags & materials.ShaderSky) != 0 {
			v = float64(c.Y) / (float64(c.ScreenHeight) - 1)

			if (stage.Flags & materials.ShaderStaticBackground) != 0 {
				u = float64(c.X) / (float64(c.ScreenWidth) - 1)
			} else {
				u = c.Angle / (2.0 * math.Pi)
			}
		}
	}

	if stage == nil || (stage.Flags&materials.ShaderTiled) != 0 {
		u *= scaleDivisor
		v *= scaleDivisor
		if stage != nil && (stage.Flags&materials.ShaderLiquid) != 0 {
			lv, lu := math.Sincos(float64(c.Frame) * constants.LiquidChurnSpeed * concepts.Deg2rad)
			u += lu * constants.LiquidChurnSize
			v += lv * constants.LiquidChurnSize
		}

		u -= math.Floor(u)
		v -= math.Floor(v)
	}

	var sample concepts.Vector4
	if image := materials.ImageFromDb(material); image != nil {
		sample = image.Sample(u, v, scale)
	} else if text := materials.TextFromDb(material); text != nil {
		sample = text.Sample(u, v, scale)
	} else if solid := materials.SolidFromDb(material); solid != nil {
		sample = solid.Diffuse.Render
	} else {
		sample[0] = 0.5
		sample[1] = 0
		sample[2] = 0.5
		sample[3] = 1.0
	}
	result.AddPreMulColorSelf(&sample)

	return result
}

func (c *Column) SampleLight(result *concepts.Vector4, material *concepts.EntityRef, world *concepts.Vector3, u, v, dist float64) *concepts.Vector4 {
	lit := materials.LitFromDb(material)

	if lit == nil {
		return result
	}

	/*// testing...
	result[0] = concepts.Clamp(dist/500.0, 0.0, 1.0)
	result[1] = result[0]
	result[2] = result[0]
	return result*/

	// Don't filter far away lightmaps. Tolerate a ~5px snap-in
	if dist > float64(c.ScreenWidth)*constants.LightGrid*0.2 {
		c.LightUnfiltered(&c.Light, world, u, v)
		return lit.Apply(result, &c.Light)
	}
	var wu, wv float64

	le00 := &c.LightElements[0]
	le10 := &c.LightElements[1]
	le11 := &c.LightElements[2]
	le01 := &c.LightElements[3]

	if c.Segment != nil && c.Normal[2] == 0 {
		le00.Type = LightElementWall
		le10.Type = LightElementWall
		le11.Type = LightElementWall
		le01.Type = LightElementWall
		c.Lightmap = c.Segment.Lightmap
		c.LightmapAge = c.Segment.LightmapAge
		le00.MapIndex = c.Segment.LightmapAddress(u, v)
		le10.MapIndex = le00.MapIndex + 1
		le11.MapIndex = le10.MapIndex + c.Segment.LightmapWidth
		le01.MapIndex = le11.MapIndex - 1
		wu = u * float64(c.Segment.LightmapWidth-constants.LightSafety*2)
		wv = v * float64(c.Segment.LightmapHeight-constants.LightSafety*2)
		wu = 1.0 - (wu - math.Floor(wu))
		wv = 1.0 - (wv - math.Floor(wv))
	} else {
		le00.Type = LightElementPlane
		le10.Type = LightElementPlane
		le11.Type = LightElementPlane
		le01.Type = LightElementPlane
		if c.Normal[2] < 0 {
			c.Lightmap = c.Sector.CeilLightmap
			c.LightmapAge = c.Sector.CeilLightmapAge
		} else {
			c.Lightmap = c.Sector.FloorLightmap
			c.LightmapAge = c.Sector.FloorLightmapAge
		}
		le00.MapIndex = c.Sector.LightmapAddress(world.To2D())
		le10.MapIndex = le00.MapIndex + 1
		le11.MapIndex = le10.MapIndex + c.Sector.LightmapWidth
		le01.MapIndex = le11.MapIndex - 1
		q := &concepts.Vector3{world[0], world[1], world[2]}
		c.Sector.ToLightmapWorld(q, c.Normal[2] > 0)
		wu = 1.0 - (world[0]-q[0])/constants.LightGrid
		wv = 1.0 - (world[1]-q[1])/constants.LightGrid
	}

	r00 := le00.Get()
	r10 := le10.Get()
	r11 := le11.Get()
	r01 := le01.Get()
	c.Light[0] = r00[0]*(wu*wv) + r10[0]*((1.0-wu)*wv) + r11[0]*(1.0-wu)*(1.0-wv) + r01[0]*wu*(1.0-wv)
	c.Light[1] = r00[1]*(wu*wv) + r10[1]*((1.0-wu)*wv) + r11[1]*(1.0-wu)*(1.0-wv) + r01[1]*wu*(1.0-wv)
	c.Light[2] = r00[2]*(wu*wv) + r10[2]*((1.0-wu)*wv) + r11[2]*(1.0-wu)*(1.0-wv) + r01[2]*wu*(1.0-wv)
	c.Light[3] = 1

	return lit.Apply(result, &c.Light)
}

func (c *Column) LightUnfiltered(result *concepts.Vector4, world *concepts.Vector3, u, v float64) *concepts.Vector4 {
	le := &c.LightElements[0]

	if c.Segment != nil && c.Normal[2] == 0 {
		le.Type = LightElementWall
		c.Lightmap = c.Segment.Lightmap
		c.LightmapAge = c.Segment.LightmapAge
		le.MapIndex = c.Segment.LightmapAddress(u, v)
	} else {
		le.Type = LightElementPlane
		if c.Normal[2] < 0 {
			c.Lightmap = c.Sector.CeilLightmap
			c.LightmapAge = c.Sector.CeilLightmapAge
		} else {
			c.Lightmap = c.Sector.FloorLightmap
			c.LightmapAge = c.Sector.FloorLightmapAge
		}
		le.MapIndex = c.Sector.LightmapAddress(world.To2D())
	}

	r00 := le.Get()
	result[0] = r00[0]
	result[1] = r00[1]
	result[2] = r00[2]
	result[3] = 1
	return result
}

func (c *Column) ApplySample(sample *concepts.Vector4, screenIndex int, z float64) {
	sample.ClampSelf(0, 1)
	if sample[3] == 0 {
		return
	}
	if sample[3] == 1 {
		c.FrameBuffer[screenIndex] = *sample
		c.ZBuffer[screenIndex] = z
		return
	}
	dst := &c.FrameBuffer[screenIndex]
	dst[0] = dst[0]*(1.0-sample[3]) + sample[0]
	dst[1] = dst[1]*(1.0-sample[3]) + sample[1]
	dst[2] = dst[2]*(1.0-sample[3]) + sample[2]
	if sample[3] > 0.8 {
		c.ZBuffer[screenIndex] = z
	}
}

func WeightBlendedOIT(c *concepts.Vector4, z float64) float64 {
	w := c[0]
	if c[1] > c[0] {
		w = c[1]
	}
	if c[2] > c[0] {
		w = c[2]
	}
	w = concepts.Clamp(w*c[3], c[3], 1.0)
	w *= concepts.Clamp(0.03/(1e-5+math.Pow(z/500, 4.0)), 1e-2, 3e3)
	return w
}
