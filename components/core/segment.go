package core

import (
	"math"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
)

type Segment struct {
	// TODO: These should probably be SimVariables
	A              *concepts.Vector2 `editable:"A"`
	B              *concepts.Vector2 `editable:"B"`
	Bottom         float64           `editable:"Bottom"`
	Top            float64           `editable:"Top"`
	Surface        materials.Surface `editable:"Mid Surface"`
	ContactScripts []*Script         `editable:"Contact Scripts"`

	Length         float64
	Normal         concepts.Vector2
	Lightmap       []concepts.Vector3
	LightmapAge    []int
	LightmapWidth  uint32
	LightmapHeight uint32
}

func (s *Segment) Recalculate() {
	s.Length = s.B.Sub(s.A).Length()
	s.Normal = concepts.Vector2{-(s.B[1] - s.A[1]) / s.Length, (s.B[0] - s.A[0]) / s.Length}
	s.LightmapWidth = uint32(s.Length/constants.LightGrid) + constants.LightSafety*2
	s.LightmapHeight = uint32((s.Top-s.Bottom)/constants.LightGrid) + constants.LightSafety*2
	s.Lightmap = make([]concepts.Vector3, s.LightmapWidth*s.LightmapHeight)
	s.LightmapAge = make([]int, s.LightmapWidth*s.LightmapHeight)
}

func (s *Segment) Matches(s2 *Segment) bool {
	d1 := math.Abs(s.A[0]-s2.A[0]) < matchEpsilon && math.Abs(s.B[0]-s2.B[0]) < matchEpsilon &&
		math.Abs(s.A[1]-s2.A[1]) < matchEpsilon && math.Abs(s.B[1]-s2.B[1]) < matchEpsilon

	d2 := math.Abs(s.A[0]-s2.B[0]) < matchEpsilon && math.Abs(s.B[0]-s2.A[0]) < matchEpsilon &&
		math.Abs(s.A[1]-s2.B[1]) < matchEpsilon && math.Abs(s.B[1]-s2.A[1]) < matchEpsilon

	return d1 || d2
}

func (s1 *Segment) Intersect2D(s2A, s2B, result *concepts.Vector2) bool {
	return concepts.IntersectSegments(s1.A, s1.B, s2A, s2B, result)
}

func (s1 *Segment) Intersect3D(s2A, s2B, result *concepts.Vector3) bool {
	r, s, s1dx, s1dy := concepts.IntersectSegmentsRaw(s1.A, s1.B, s2A.To2D(), s2B.To2D())
	if r < 0 {
		return false
	}
	result[0] = s1.A[0] + r*s1dx
	result[1] = s1.A[1] + r*s1dy
	result[2] = (1.0-s)*s2A[2] + s*s2B[2]
	return true
}

func (s *Segment) AABBIntersect(xMin, yMin, xMax, yMax float64) bool {
	// Find min and mA[0] X for the segment
	minX := s.A[0]
	maxX := s.B[0]

	if s.A[0] > s.B[0] {
		minX = s.B[0]
		maxX = s.A[0]
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

	// Find corresponding min and mA[0] Y for min and mA[0] X we found before
	minY := s.A[1]
	maxY := s.B[1]
	dx := s.B[0] - s.A[0]

	if math.Abs(dx) > constants.IntersectEpsilon {
		a := (s.B[1] - s.A[1]) / dx
		b := s.A[1] - a*s.A[0]
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

func (s *Segment) DistanceToPoint2(p *concepts.Vector2) float64 {
	l2 := s.A.Dist2(s.B)
	if l2 == 0 {
		return p.Dist2(s.A)
	}
	delta := &concepts.Vector2{s.B[0] - s.A[0], s.B[1] - s.A[1]}
	t := (&concepts.Vector2{p[0], p[1]}).SubSelf(s.A).Dot(delta) / l2
	if t < 0 {
		return p.Dist2(s.A)
	}
	if t > 1 {
		return p.Dist2(s.B)
	}
	return p.Dist2(delta.MulSelf(t).AddSelf(s.A))
}

func (s *Segment) DistanceToPoint(p *concepts.Vector2) float64 {
	return math.Sqrt(s.DistanceToPoint2(p))
}

func (s *Segment) ClosestToPoint(p *concepts.Vector2) *concepts.Vector2 {
	delta := s.B.Sub(s.A)
	dist2 := delta[0]*delta[0] + delta[1]*delta[1]
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
	return s.A.Add(delta.MulSelf(t))
}

func (s *Segment) WhichSide(p *concepts.Vector2) float64 {
	return s.Normal.Dot(p.Sub(s.A))
}

func (s *Segment) LightmapAddress(u, v float64) uint32 {
	dx := int(u*float64(s.LightmapWidth-constants.LightSafety*2)) + constants.LightSafety
	dy := int(v*float64(s.LightmapHeight-constants.LightSafety*2)) + constants.LightSafety
	return uint32(dy)*s.LightmapWidth + uint32(dx)
}

func (s *Segment) UVToWorld(result *concepts.Vector3, u, v float64) *concepts.Vector3 {
	result[0] = s.A[0] + (s.B[0]-s.A[0])*u
	result[1] = s.A[1] + (s.B[1]-s.A[1])*u
	// v goes from top (0) to bottom (1)
	result[2] = v*s.Bottom + (1.0-v)*s.Top
	return result
}

func (s *Segment) LightmapAddressToWorld(result *concepts.Vector3, mapIndex uint32) *concepts.Vector3 {
	lu := (mapIndex % s.LightmapWidth) - constants.LightSafety
	lv := (mapIndex / s.LightmapWidth) - constants.LightSafety
	u := (float64(lu) + 0.5) / float64(s.LightmapWidth-(constants.LightSafety*2))
	v := (float64(lv) + 0.5) / float64(s.LightmapHeight-(constants.LightSafety*2))
	return s.UVToWorld(result, u, v)
}

func (s *Segment) Construct(db *concepts.EntityComponentDB, data map[string]any) {
	s.A = new(concepts.Vector2)
	s.B = new(concepts.Vector2)
	s.Normal = concepts.Vector2{}
	s.Bottom = 0
	s.Top = 64

	if data == nil {
		return
	}

	if v, ok := data["A"]; ok {
		s.A.Deserialize(v.(map[string]any))
	}
	if v, ok := data["B"]; ok {
		s.B.Deserialize(v.(map[string]any))
	}
	if v, ok := data["Top"]; ok {
		s.Top = v.(float64)
	}
	if v, ok := data["Bottom"]; ok {
		s.Bottom = v.(float64)
	}
	if v, ok := data["Mid"]; ok {
		s.Surface.Construct(db, v.(map[string]any))
	}
	if v, ok := data["Surface"]; ok {
		s.Surface.Construct(db, v.(map[string]any))
	}
	if v, ok := data["ContactScripts"]; ok {
		s.ContactScripts = concepts.ConstructSlice[*Script](db, v)
	}
}

func (s *Segment) Serialize(storePositions bool) map[string]any {
	result := make(map[string]any)
	if storePositions {
		result["A"] = s.A.Serialize()
		result["B"] = s.B.Serialize()
		result["Top"] = s.Top
		result["Bottom"] = s.Bottom
	}
	result["Surface"] = s.Surface.Serialize()

	if len(s.ContactScripts) > 0 {
		result["ContactScripts"] = concepts.SerializeSlice(s.ContactScripts)
	}

	return result
}
