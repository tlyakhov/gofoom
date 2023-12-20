package core

import (
	"math"

	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
)

const (
	matchEpsilon float64 = 1e-4
)

type Segment struct {
	DB *concepts.EntityComponentDB

	P                 concepts.Vector2  `editable:"X/Y"`
	LoSurface         materials.Surface `editable:"Low Surface"`
	MidSurface        materials.Surface `editable:"Mid Surface"`
	HiSurface         materials.Surface `editable:"High Surface"`
	PortalHasMaterial bool              `editable:"Portal has material"`
	PortalIsPassable  bool              `editable:"Portal is passable"`
	ContactScripts    []Script          `editable:"Contact Scripts"`

	AdjacentSector  *concepts.EntityRef
	AdjacentSegment *Segment

	Length         float64
	Normal         concepts.Vector2
	Sector         *Sector
	Next           *Segment
	Prev           *Segment
	Lightmap       []concepts.Vector3
	LightmapAge    []int
	LightmapWidth  uint32
	LightmapHeight uint32
	Flags          int
}

var SegmentComponentIndex int

func init() {
	SegmentComponentIndex = concepts.DbTypes().Register(Segment{}, SectorFromDb)
}

func (s *Segment) RealizeAdjacentSector() {
	if s.AdjacentSector.Nil() {
		return
	}

	if adj, ok := s.AdjacentSector.Component(SectorComponentIndex).(*Sector); ok {
		// Get the actual segment
		for _, s2 := range adj.Segments {
			if s2.Matches(s) {
				s.AdjacentSegment = s2
				break
			}
		}
	}
}

func (s *Segment) Recalculate() {
	s.Length = s.Next.P.Sub(&s.P).Length()
	s.Normal = concepts.Vector2{-(s.Next.P[1] - s.P[1]) / s.Length, (s.Next.P[0] - s.P[0]) / s.Length}
	if s.Sector != nil {
		s.RealizeAdjacentSector()
		s.LightmapWidth = uint32(s.Length/constants.LightGrid) + constants.LightSafety*2
		s.LightmapHeight = uint32((s.Sector.Max[2]-s.Sector.Min[2])/constants.LightGrid) + constants.LightSafety*2
		s.Lightmap = make([]concepts.Vector3, s.LightmapWidth*s.LightmapHeight)
		s.LightmapAge = make([]int, s.LightmapWidth*s.LightmapHeight)
	}
}

func (s *Segment) Matches(s2 *Segment) bool {
	d1 := math.Abs(s.P[0]-s2.P[0]) < matchEpsilon && math.Abs(s.Next.P[0]-s2.Next.P[0]) < matchEpsilon &&
		math.Abs(s.P[1]-s2.P[1]) < matchEpsilon && math.Abs(s.Next.P[1]-s2.Next.P[1]) < matchEpsilon

	d2 := math.Abs(s.P[0]-s2.Next.P[0]) < matchEpsilon && math.Abs(s.Next.P[0]-s2.P[0]) < matchEpsilon &&
		math.Abs(s.P[1]-s2.Next.P[1]) < matchEpsilon && math.Abs(s.Next.P[1]-s2.P[1]) < matchEpsilon

	return d1 || d2
}

func (s1 *Segment) intersect(s2A, s2B *concepts.Vector2) (float64, float64, float64, float64) {
	s1dx := s1.Next.P[0] - s1.P[0]
	s1dy := s1.Next.P[1] - s1.P[1]
	s2dx := s2B[0] - s2A[0]
	s2dy := s2B[1] - s2A[1]

	denom := s1dx*s2dy - s2dx*s1dy
	if denom == 0 {
		return -1, -1, -1, -1
	}
	r := (s1.P[1]-s2A[1])*s2dx - (s1.P[0]-s2A[0])*s2dy
	if (denom < 0 && r >= constants.IntersectEpsilon) ||
		(denom > 0 && r < -constants.IntersectEpsilon) {
		return -1, -1, -1, -1
	}
	s := (s1.P[1]-s2A[1])*s1dx - (s1.P[0]-s2A[0])*s1dy
	if (denom < 0 && s >= constants.IntersectEpsilon) ||
		(denom > 0 && s < -constants.IntersectEpsilon) {
		return -1, -1, -1, -1
	}
	r /= denom
	s /= denom
	if r > 1.0+constants.IntersectEpsilon || s > 1.0+constants.IntersectEpsilon {
		return -1, -1, -1, -1
	}
	return concepts.Clamp(r, 0.0, 1.0), concepts.Clamp(s, 0.0, 1.0), s1dx, s1dy
}

func (s1 *Segment) Intersect2D(s2A, s2B, result *concepts.Vector2) bool {
	r, _, s1dx, s1dy := s1.intersect(s2A, s2B)
	if r < 0 {
		return false
	}
	result[0] = s1.P[0] + r*s1dx
	result[1] = s1.P[1] + r*s1dy
	return true
}

func (s1 *Segment) Intersect3D(s2A, s2B, result *concepts.Vector3) bool {
	r, s, s1dx, s1dy := s1.intersect(s2A.To2D(), s2B.To2D())
	if r < 0 {
		return false
	}
	result[0] = s1.P[0] + r*s1dx
	result[1] = s1.P[1] + r*s1dy
	result[2] = (1.0-s)*s2A[2] + s*s2B[2]
	return true
}

func (s *Segment) AABBIntersect(xMin, yMin, xMax, yMax float64) bool {
	// Find min and mA[0] X for the segment
	minX := s.P[0]
	maxX := s.Next.P[0]

	if s.P[0] > s.Next.P[0] {
		minX = s.Next.P[0]
		maxX = s.P[0]
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
	minY := s.P[1]
	maxY := s.Next.P[1]
	dx := s.Next.P[0] - s.P[0]

	if math.Abs(dx) > constants.IntersectEpsilon {
		a := (s.Next.P[1] - s.P[1]) / dx
		b := s.P[1] - a*s.P[0]
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
	l2 := s.P.Dist2(&s.Next.P)
	if l2 == 0 {
		return p.Dist2(&s.P)
	}
	delta := &concepts.Vector2{s.Next.P[0] - s.P[0], s.Next.P[1] - s.P[1]}
	t := (&concepts.Vector2{p[0], p[1]}).SubSelf(&s.P).Dot(delta) / l2
	if t < 0 {
		return p.Dist2(&s.P)
	}
	if t > 1 {
		return p.Dist2(&s.Next.P)
	}
	return p.Dist2(delta.MulSelf(t).AddSelf(&s.P))
}

func (s *Segment) DistanceToPoint(p *concepts.Vector2) float64 {
	return math.Sqrt(s.DistanceToPoint2(p))
}

func (s *Segment) ClosestToPoint(p *concepts.Vector2) *concepts.Vector2 {
	delta := s.Next.P.Sub(&s.P)
	dist2 := delta[0]*delta[0] + delta[1]*delta[1]
	if dist2 == 0 {
		return &s.P
	}
	ap := p.Sub(&s.P)
	t := ap.Dot(delta) / dist2

	if t < 0 {
		return &s.P
	}
	if t > 1 {
		return &s.Next.P
	}
	return s.P.Add(delta.MulSelf(t))
}

func (s *Segment) WhichSide(p *concepts.Vector2) float64 {
	return s.Normal.Dot(p.Sub(&s.P))
}

func (s *Segment) LightmapAddress(u, v float64) uint32 {
	dx := int(u*float64(s.LightmapWidth-constants.LightSafety*2)) + constants.LightSafety
	dy := int(v*float64(s.LightmapHeight-constants.LightSafety*2)) + constants.LightSafety
	return uint32(dy)*s.LightmapWidth + uint32(dx)
}

func (s *Segment) UVToWorld(result *concepts.Vector3, u, v float64) *concepts.Vector3 {
	result[0] = s.P[0] + (s.Next.P[0]-s.P[0])*u
	result[1] = s.P[1] + (s.Next.P[1]-s.P[1])*u
	// v goes from top (0) to bottom (1)
	result[2] = v*s.Sector.Min[2] + (1.0-v)*s.Sector.Max[2]
	return result
}

func (s *Segment) LightmapAddressToWorld(result *concepts.Vector3, mapIndex uint32) *concepts.Vector3 {
	lu := (mapIndex % s.LightmapWidth) - constants.LightSafety
	lv := (mapIndex / s.LightmapWidth) - constants.LightSafety
	u := (float64(lu) + 0.5) / float64(s.LightmapWidth-(constants.LightSafety*2))
	v := (float64(lv) + 0.5) / float64(s.LightmapHeight-(constants.LightSafety*2))
	return s.UVToWorld(result, u, v)
}

func (s *Segment) Split(p concepts.Vector2) *Segment {
	// Segments are linked list where the line goes from `s.P` -> `s.Next.P`
	// The insertion index for the split should therefore be such that `s` is
	// the _previous_ segment
	index := 0
	for index < len(s.Sector.Segments) &&
		s.Sector.Segments[index].Prev != s {
		index++
	}
	// Make a copy to preserve everything other than the point.
	copied := new(Segment)
	copied.Sector = s.Sector
	copied.DB = s.DB
	copied.Construct(s.Serialize())
	copied.P = p
	// Insert into the sector
	s.Sector.Segments = append(s.Sector.Segments[:index+1], s.Sector.Segments[index:]...)
	s.Sector.Segments[index] = copied
	// Recalculate metadata
	//DbgPrintSegments(seg.Sector)
	s.Sector.Recalculate()
	//DbgPrintSegments(seg.Sector)
	return copied
}

func (s *Segment) Construct(data map[string]any) {
	s.P = concepts.Vector2{}
	s.Normal = concepts.Vector2{}
	s.PortalHasMaterial = false
	s.PortalIsPassable = true

	if data == nil {
		return
	}

	if v, ok := data["X"]; ok {
		s.P[0] = v.(float64)
	}
	if v, ok := data["Y"]; ok {
		s.P[1] = v.(float64)
	}
	if v, ok := data["PortalHasMaterial"]; ok {
		s.PortalHasMaterial = v.(bool)
	}
	if v, ok := data["PortalIsPassable"]; ok {
		s.PortalIsPassable = v.(bool)
	}

	if v, ok := data["AdjacentSector"]; ok {
		s.AdjacentSector = s.DB.DeserializeEntityRef(v)
	}
	if v, ok := data["Lo"]; ok {
		s.LoSurface.Construct(s.DB, v.(map[string]any))
	}
	if v, ok := data["Mid"]; ok {
		s.MidSurface.Construct(s.DB, v.(map[string]any))
	}
	if v, ok := data["Hi"]; ok {
		s.HiSurface.Construct(s.DB, v.(map[string]any))
	}
	if v, ok := data["ContactScripts"]; ok {
		s.ContactScripts = ConstructScripts(s.DB, v)
	}
}

func (s *Segment) Serialize() map[string]any {
	result := make(map[string]any)
	result["X"] = s.P[0]
	result["Y"] = s.P[1]

	if s.PortalHasMaterial {
		result["PortalHasMaterial"] = true
	}
	if !s.PortalIsPassable {
		result["PortalIsPassable"] = false
	}

	result["Lo"] = s.LoSurface.Serialize()
	result["Mid"] = s.MidSurface.Serialize()
	result["Hi"] = s.HiSurface.Serialize()

	if !s.AdjacentSector.Nil() {
		result["AdjacentSector"] = s.AdjacentSector.Serialize()
	}
	if len(s.ContactScripts) > 0 {
		result["ContactScripts"] = SerializeScripts(s.ContactScripts)
	}

	return result
}
