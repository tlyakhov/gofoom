package state

import (
	"math"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
)

type Ray struct {
	Start, End concepts.Vector2
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
	LightElements      [4]LightElement
	Normal             concepts.Vector3
	Material           concepts.Vector4
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

func (c *Column) SampleMaterial(material *concepts.EntityRef, u, v float64, scale float64) *concepts.Vector4 {
	// Should refactor this scale thing, it's weird
	scaleDivisor := 1.0

	if tiled := materials.TiledFromDb(material); tiled != nil {
		scaleDivisor *= tiled.Scale
		u *= scaleDivisor
		v *= scaleDivisor
		if tiled.IsLiquid {
			u += math.Cos(float64(c.Frame)*constants.LiquidChurnSpeed*concepts.Deg2rad) * constants.LiquidChurnSize
			v += math.Sin(float64(c.Frame)*constants.LiquidChurnSpeed*concepts.Deg2rad) * constants.LiquidChurnSize
		}

		u -= math.Floor(u)
		v -= math.Floor(v)
		u = math.Abs(u)
		v = math.Abs(v)
	}

	if sky := materials.SkyFromDb(material); sky != nil {
		v = float64(c.Y) / (float64(c.ScreenHeight) - 1)

		if sky.StaticBackground {
			u = float64(c.X) / (float64(c.ScreenWidth) - 1)
		} else {
			u = c.Angle / (2.0 * math.Pi)
			for ; u < 0; u++ {
			}
			for ; u > 1; u-- {
			}
		}
	}

	if image := materials.ImageFromDb(material); image != nil {
		c.Material = image.Sample(u, v, scale)
	} else if solid := materials.SolidFromDb(material); solid != nil {
		c.Material[0] = float64(solid.Diffuse.R) / 255.0
		c.Material[1] = float64(solid.Diffuse.G) / 255.0
		c.Material[2] = float64(solid.Diffuse.B) / 255.0
	} else {
		c.Material[0] = 0.5
		c.Material[1] = 0.5
		c.Material[2] = 0.5
		c.Material[3] = 1.0
	}

	return &c.Material
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
