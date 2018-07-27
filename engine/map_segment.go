package engine

import (
	"math"

	"github.com/tlyakhov/gofoom/constants"

	"github.com/tlyakhov/gofoom/util"
)

type MaterialBehavior int

//go:generate stringer -type=MaterialBehavior
const (
	ScaleNone MaterialBehavior = iota
	ScaleWidth
	ScaleHeight
	ScaleAll
)

const (
	matchEpsilon float64 = 1e-4
)

type MapSegment struct {
	util.CommonFields

	A, B            *util.Vector2
	LoMaterial      *Material
	MidMaterial     *Material
	HiMaterial      *Material
	LoBehavior      MaterialBehavior
	Midhavior       MaterialBehavior
	HiBehavior      MaterialBehavior
	Length          float64
	Normal          *util.Vector2
	Sector          *MapSector
	AdjacentSector  *MapSector
	AdjacentSegment *MapSegment
	Lightmap        []float64
	LightmapWidth   uint
	LightmapHeight  uint
	Flags           int
}

func (ms *MapSegment) Update() {
	ms.Length = ms.B.Sub(ms.A).Length()
	ms.Normal = &util.Vector2{-(ms.B.Y - ms.A.Y) / ms.Length, (ms.B.X - ms.A.X) / ms.Length}
	if ms.Sector != nil {
		ms.LightmapWidth = uint(ms.Length/constants.LightGrid) + 2
		ms.LightmapHeight = uint((ms.Sector.TopZ-ms.Sector.BottomZ)/constants.LightGrid) + 2
		ms.Lightmap = make([]float64, ms.LightmapWidth*ms.LightmapHeight*3)
		ms.ClearLightmap()
	}
}

func (ms *MapSegment) ClearLightmap() {
	for i := range ms.Lightmap {
		ms.Lightmap[i] = -1
	}
}

func (ms *MapSegment) Matches(s2 *MapSegment) bool {
	d1 := math.Abs(ms.A.X-s2.A.X) < matchEpsilon && math.Abs(ms.B.X-s2.B.X) < matchEpsilon &&
		math.Abs(ms.A.Y-s2.A.Y) < matchEpsilon && math.Abs(ms.B.Y-s2.B.Y) < matchEpsilon

	d2 := math.Abs(ms.A.X-s2.B.X) < matchEpsilon && math.Abs(ms.B.X-s2.A.X) < matchEpsilon &&
		math.Abs(ms.A.Y-s2.B.Y) < matchEpsilon && math.Abs(ms.B.Y-s2.A.Y) < matchEpsilon

	return d1 || d2
}

func (s1 *MapSegment) Intersect(s2A, s2B *util.Vector2) *util.Vector2 {
	s1dx := s1.B.X - s1.A.X
	s1dy := s1.B.Y - s1.A.Y
	s2dx := s2B.X - s2A.X
	s2dy := s2B.Y - s2A.Y

	denom := s1dx*s2dy - s2dx*s1dy
	if denom == 0 {
		return nil
	}
	r := (s1.A.Y-s2A.Y)*s2dx - (s1.A.X-s2A.X)*s2dy
	if (denom < 0 && r >= constants.IntersectEpsilon) ||
		(denom > 0 && r < -constants.IntersectEpsilon) {
		return nil
	}
	s := (s1.A.Y-s2A.Y)*s1dx - (s1.A.X-s2A.X)*s1dy
	if (denom < 0 && s >= constants.IntersectEpsilon) ||
		(denom > 0 && s < -constants.IntersectEpsilon) {
		return nil
	}
	r /= denom
	s /= denom
	if r > 1.0+constants.IntersectEpsilon || s > 1.0+constants.IntersectEpsilon {
		return nil
	}
	r = util.Clamp(r, 0.0, 1.0)
	return &util.Vector2{s1.A.X + r*s1dx, s1.A.Y + r*s1dy}
}

func (ms *MapSegment) AABBIntersect(xMin, yMin, xMax, yMax float64) bool {
	// Find min and mA.X X for the segment
	minX := ms.A.X
	maxX := ms.B.X

	if ms.A.X > ms.B.X {
		minX = ms.B.X
		maxX = ms.A.X
	}

	// Find the intersection of the segment's and rectangle's x-projections
	if maxX > xMax {
		maxX = xMax
	}
	if minX < xMin {
		minX = xMin
	}
	// If their projections do not intersect return false
	if minX > maxX {
		return false
	}

	// Find corresponding min and mA.X Y for min and mA.X X we found before
	minY := ms.A.Y
	maxY := ms.B.Y
	dx := ms.B.X - ms.A.X

	if math.Abs(dx) > constants.IntersectEpsilon {
		a := (ms.B.Y - ms.A.Y) / dx
		b := ms.A.Y - a*ms.A.X
		minY = a*minX + b
		maxY = a*maxX + b
	}
	if minY > maxY {
		tmp := maxY
		maxY = minY
		minY = tmp
	}

	// Find the intersection of the segment's and rectangle's y-projections
	if maxY > yMax {
		maxY = yMax
	}
	if minY < yMin {
		minY = yMin
	}

	return minY <= maxY // If Y-projections do not intersect return false
}

func (ms *MapSegment) DistanceToPoint2(p *util.Vector2) float64 {
	l2 := ms.A.Dist2(ms.B)
	if l2 == 0 {
		return p.Dist2(ms.A)
	}
	t := p.Sub(ms.A).Dot(ms.B.Sub(ms.A))
	if t < 0 {
		return p.Dist2(ms.A)
	}
	if t > 1 {
		return p.Dist2(ms.B)
	}
	return p.Dist2(ms.A.Add(ms.B.Sub(ms.A).Mul(t)))
}

func (ms *MapSegment) DistanceToPoint(p *util.Vector2) float64 {
	return math.Sqrt(ms.DistanceToPoint2(p))
}

func (ms *MapSegment) ClosestToPoint(p *util.Vector2) *util.Vector2 {
	delta := ms.B.Sub(ms.A)
	dist2 := delta.X*delta.X + delta.Y*delta.Y
	if dist2 == 0 {
		return ms.A
	}
	ap := p.Sub(ms.A)
	t := ap.Dot(delta) / dist2

	if t < 0 {
		return ms.A
	}
	if t > 1 {
		return ms.B
	}
	return ms.A.Add(delta.Mul(t))
}

func (ms *MapSegment) WhichSide(p *util.Vector2) float64 {
	return ms.Normal.Dot(p.Sub(ms.A))
}

func (ms *MapSegment) UVToWorld(u, v float64) *util.Vector3 {
	alongSegment := ms.A.Add(ms.B.Sub(ms.A).Mul(u))
	return &util.Vector3{alongSegment.X, alongSegment.Y, v*ms.Sector.BottomZ + (1.0-v)*ms.Sector.TopZ}
}

func (ms *MapSegment) LMAddressToWorld(mapIndex uint) *util.Vector3 {
	lu := ((mapIndex / 3) % ms.LightmapWidth) - 1
	lv := ((mapIndex / 3) / ms.LightmapWidth) - 1
	u := float64(lu) / float64(ms.LightmapWidth-2)
	v := float64(lv) / float64(ms.LightmapHeight-2)
	return ms.UVToWorld(u, v)
}
