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
	PickInternalSegment
)

type PickedElement struct {
	Type    PickedType // ceil, floor, mid, hi, lo, body
	Element any
}

type Column struct {
	*Config
	MaterialSampler
	X, Y, YStart, YEnd int
	Sector             *core.Sector
	Segment            *core.Segment
	SectorSegment      *core.SectorSegment
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
	BottomZ            float64
	TopZ               float64
	ProjHeightTop      float64
	ProjHeightBottom   float64
	ScreenStart        int
	ScreenEnd          int
	ClippedStart       int
	ClippedEnd         int
	LightElements      [4]LightElement
	Light              concepts.Vector4
	Pick               bool
	PickedElements     []PickedElement
}

func (c *Column) IntersectSegment(segment *core.Segment, checkDist bool) bool {
	// Wall is facing away from us
	if c.Ray.Delta.Dot(&segment.Normal) > 0 {
		return false
	}

	// Ray intersects?
	isect := c.Intersection.To2D()
	if ok := segment.Intersect2D(&c.Ray.Start, &c.Ray.End, isect); !ok {
		/*	if c.Sector.Entity == 82 {
			dbg := fmt.Sprintf("No intersection %v <-> %v", segment.A.StringHuman(), segment.B.StringHuman())
			c.DebugNotices.Push(dbg)
		}*/
		return false
	}

	var dist float64
	delta := concepts.Vector2{math.Abs(c.Intersection[0] - c.Ray.Start[0]), math.Abs(c.Intersection[1] - c.Ray.Start[1])}
	if delta[1] > delta[0] {
		dist = math.Abs(delta[1] / c.AngleSin)
	} else {
		dist = math.Abs(delta[0] / c.AngleCos)
	}

	if checkDist && (dist > c.Distance || dist < c.LastPortalDistance) {
		return false
	}

	c.Segment = segment
	c.Distance = dist
	c.U = isect.Dist(segment.A) / segment.Length
	return true
}

func (c *Column) ProjectZ(z float64) float64 {
	return z * c.ViewFix[c.X] / c.Distance
}

func (c *Column) CalcScreen() {
	// Screen slice precalculation
	c.ProjHeightTop = c.ProjectZ(c.TopZ - c.CameraZ)
	c.ProjHeightBottom = c.ProjectZ(c.BottomZ - c.CameraZ)

	c.ScreenStart = c.ScreenHeight/2 - int(c.ProjHeightTop)
	c.ScreenEnd = c.ScreenHeight/2 - int(c.ProjHeightBottom)
	c.ClippedStart = concepts.IntClamp(c.ScreenStart, c.YStart, c.YEnd)
	c.ClippedEnd = concepts.IntClamp(c.ScreenEnd, c.YStart, c.YEnd)
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

	if c.Segment != nil && le00.Type == LightElementWall {
		le00.MapIndex = c.Segment.LightmapAddress(u, v)
		le10.MapIndex = le00.MapIndex + 1
		le11.MapIndex = le10.MapIndex + c.Segment.LightmapWidth
		le01.MapIndex = le11.MapIndex - 1
		wu = u * float64(c.Segment.LightmapWidth-constants.LightSafety*2)
		wv = v * float64(c.Segment.LightmapHeight-constants.LightSafety*2)
		wu = 1.0 - (wu - math.Floor(wu))
		wv = 1.0 - (wv - math.Floor(wv))
	} else {
		le00.MapIndex = c.Sector.LightmapAddress(world.To2D())
		le10.MapIndex = le00.MapIndex + 1
		le11.MapIndex = le10.MapIndex + c.Sector.LightmapWidth
		le01.MapIndex = le11.MapIndex - 1
		q := &concepts.Vector3{world[0], world[1], world[2]}
		c.Sector.ToLightmapWorld(q, le00.Type == LightElementWall)
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

	if c.Segment != nil && le.Type == LightElementWall {
		le.MapIndex = c.Segment.LightmapAddress(u, v)
	} else {
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
