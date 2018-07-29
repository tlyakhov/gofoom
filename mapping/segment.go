package mapping

import (
	"math"

	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/constants"
	fmath "github.com/tlyakhov/gofoom/math"
)

const (
	matchEpsilon float64 = 1e-4
)

type Segment struct {
	concepts.Base

	A, B            *fmath.Vector2
	LoMaterial      concepts.ISerializable
	MidMaterial     concepts.ISerializable
	HiMaterial      concepts.ISerializable
	LoBehavior      MaterialBehavior
	Midhavior       MaterialBehavior
	HiBehavior      MaterialBehavior
	Length          float64
	Normal          *fmath.Vector2
	Sector          *Sector
	AdjacentSector  *Sector
	AdjacentSegment *Segment
	Lightmap        []float64
	LightmapWidth   uint
	LightmapHeight  uint
	Flags           int
}

func (s *Segment) Update() {
	s.Length = s.B.Sub(s.A).Length()
	s.Normal = &fmath.Vector2{-(s.B.Y - s.A.Y) / s.Length, (s.B.X - s.A.X) / s.Length}
	if s.Sector != nil {
		s.LightmapWidth = uint(s.Length/constants.LightGrid) + 2
		s.LightmapHeight = uint((s.Sector.TopZ-s.Sector.BottomZ)/constants.LightGrid) + 2
		s.Lightmap = make([]float64, s.LightmapWidth*s.LightmapHeight*3)
		s.ClearLightmap()
	}
}

func (s *Segment) ClearLightmap() {
	for i := range s.Lightmap {
		s.Lightmap[i] = -1
	}
}

func (s *Segment) Matches(s2 *Segment) bool {
	d1 := math.Abs(s.A.X-s2.A.X) < matchEpsilon && math.Abs(s.B.X-s2.B.X) < matchEpsilon &&
		math.Abs(s.A.Y-s2.A.Y) < matchEpsilon && math.Abs(s.B.Y-s2.B.Y) < matchEpsilon

	d2 := math.Abs(s.A.X-s2.B.X) < matchEpsilon && math.Abs(s.B.X-s2.A.X) < matchEpsilon &&
		math.Abs(s.A.Y-s2.B.Y) < matchEpsilon && math.Abs(s.B.Y-s2.A.Y) < matchEpsilon

	return d1 || d2
}

func (s1 *Segment) Intersect(s2A, s2B *fmath.Vector2) *fmath.Vector2 {
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
	r = fmath.Clamp(r, 0.0, 1.0)
	return &fmath.Vector2{s1.A.X + r*s1dx, s1.A.Y + r*s1dy}
}

func (s *Segment) AABBIntersect(xMin, yMin, xMax, yMax float64) bool {
	// Find min and mA.X X for the segment
	minX := s.A.X
	maxX := s.B.X

	if s.A.X > s.B.X {
		minX = s.B.X
		maxX = s.A.X
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
	minY := s.A.Y
	maxY := s.B.Y
	dx := s.B.X - s.A.X

	if math.Abs(dx) > constants.IntersectEpsilon {
		a := (s.B.Y - s.A.Y) / dx
		b := s.A.Y - a*s.A.X
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

func (s *Segment) DistanceToPoint2(p *fmath.Vector2) float64 {
	l2 := s.A.Dist2(s.B)
	if l2 == 0 {
		return p.Dist2(s.A)
	}
	t := p.Sub(s.A).Dot(s.B.Sub(s.A))
	if t < 0 {
		return p.Dist2(s.A)
	}
	if t > 1 {
		return p.Dist2(s.B)
	}
	return p.Dist2(s.A.Add(s.B.Sub(s.A).Mul(t)))
}

func (s *Segment) DistanceToPoint(p *fmath.Vector2) float64 {
	return math.Sqrt(s.DistanceToPoint2(p))
}

func (s *Segment) ClosestToPoint(p *fmath.Vector2) *fmath.Vector2 {
	delta := s.B.Sub(s.A)
	dist2 := delta.X*delta.X + delta.Y*delta.Y
	if dist2 == 0 {
		return s.A
	}
	ap := p.Sub(s.A)
	t := ap.Dot(delta) / dist2

	if t < 0 {
		return s.A
	}
	if t > 1 {
		return s.B
	}
	return s.A.Add(delta.Mul(t))
}

func (s *Segment) WhichSide(p *fmath.Vector2) float64 {
	return s.Normal.Dot(p.Sub(s.A))
}

func (s *Segment) UVToWorld(u, v float64) *fmath.Vector3 {
	alongSegment := s.A.Add(s.B.Sub(s.A).Mul(u))
	return &fmath.Vector3{alongSegment.X, alongSegment.Y, v*s.Sector.BottomZ + (1.0-v)*s.Sector.TopZ}
}

func (s *Segment) LMAddressToWorld(mapIndex uint) *fmath.Vector3 {
	lu := ((mapIndex / 3) % s.LightmapWidth) - 1
	lv := ((mapIndex / 3) / s.LightmapWidth) - 1
	u := float64(lu) / float64(s.LightmapWidth-2)
	v := float64(lv) / float64(s.LightmapHeight-2)
	return s.UVToWorld(u, v)
}
