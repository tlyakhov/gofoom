package core

import (
	"math"

	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/registry"
)

const (
	matchEpsilon float64 = 1e-4
)

type Segment struct {
	concepts.Base `editable:"^"`

	P           concepts.Vector2       `editable:"X/Y"`
	LoMaterial  concepts.ISerializable `editable:"Low Material" edit_type:"Material"`
	MidMaterial concepts.ISerializable `editable:"Mid Material" edit_type:"Material"`
	HiMaterial  concepts.ISerializable `editable:"High Material" edit_type:"Material"`
	LoBehavior  MaterialBehavior       `editable:"Low Behavior"`
	MidBehavior MaterialBehavior       `editable:"Mid Behavior"`
	HiBehavior  MaterialBehavior       `editable:"High Behavior"`

	AdjacentSector  AbstractSector
	AdjacentSegment *Segment

	Length         float64
	Normal         concepts.Vector2
	Sector         AbstractSector
	Next           *Segment
	Prev           *Segment
	Lightmap       []concepts.Vector3
	LightmapAge    []int
	LightmapWidth  uint32
	LightmapHeight uint32
	Flags          int
}

func init() {
	registry.Instance().Register(Segment{})
}

func (s *Segment) Initialize() {
	s.Base.Initialize()
	s.P = concepts.Vector2{}
	s.Normal = concepts.Vector2{}
}

func (s *Segment) SetParent(parent interface{}) {
	if sector, ok := parent.(AbstractSector); ok {
		s.Sector = sector
	} else {
		panic("Tried core.Segment.SetParent with a parameter that wasn't a *core.AbstractSector")
	}
}

func (s *Segment) RealizeAdjacentSector() {
	if s.AdjacentSector == nil {
		return
	}
	if ph, ok := s.AdjacentSector.(*PlaceholderSector); ok {
		// Get the actual one.
		if adj, ok := s.Sector.Physical().Map.Sectors[ph.ID]; ok {
			s.AdjacentSector = adj
			for _, s2 := range s.AdjacentSector.Physical().Segments {
				if s2.Matches(s) {
					s.AdjacentSegment = s2
					break
				}
			}
		}
	}
}

func (s *Segment) Recalculate() {
	s.Length = s.Next.P.Sub(&s.P).Length()
	s.Normal = concepts.Vector2{-(s.Next.P[1] - s.P[1]) / s.Length, (s.Next.P[0] - s.P[0]) / s.Length}
	if s.Sector != nil {
		s.RealizeAdjacentSector()
		sector := s.Sector.Physical()
		s.LightmapWidth = uint32(s.Length/constants.LightGrid) + constants.LightSafety*2
		s.LightmapHeight = uint32((sector.TopZ-sector.BottomZ)/constants.LightGrid) + constants.LightSafety*2
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
	result[0] = s.Next.P[0]
	result[1] = s.Next.P[1]
	result[2] = v*s.Sector.Physical().BottomZ + (1.0-v)*s.Sector.Physical().TopZ
	result.To2D().SubSelf(&s.P).MulSelf(u).AddSelf(&s.P)
	return result
}

func (s *Segment) LightmapAddressToWorld(result *concepts.Vector3, mapIndex uint32) *concepts.Vector3 {
	lu := (mapIndex % s.LightmapWidth) - constants.LightSafety
	lv := (mapIndex / s.LightmapWidth) - constants.LightSafety
	u := (float64(lu) + 0.5) / float64(s.LightmapWidth-(constants.LightSafety*2))
	v := (float64(lv) + 0.5) / float64(s.LightmapHeight-(constants.LightSafety*2))
	return s.UVToWorld(result, u, v)
}

func (s *Segment) Deserialize(data map[string]interface{}) {
	s.Initialize()
	s.Base.Deserialize(data)
	if v, ok := data["X"]; ok {
		s.P[0] = v.(float64)
	}
	if v, ok := data["Y"]; ok {
		s.P[1] = v.(float64)
	}
	if v, ok := data["AdjacentSector"]; ok {
		s.AdjacentSector = &PlaceholderSector{Base: concepts.Base{ID: v.(string)}}
	}
	if v, ok := data["LoMaterial"]; ok {
		s.LoMaterial = s.Sector.Physical().Map.Materials[v.(string)]
	}
	if v, ok := data["MidMaterial"]; ok {
		s.MidMaterial = s.Sector.Physical().Map.Materials[v.(string)]
	}
	if v, ok := data["HiMaterial"]; ok {
		s.HiMaterial = s.Sector.Physical().Map.Materials[v.(string)]
	}
	if v, ok := data["LoBehavior"]; ok {
		mb, error := MaterialBehaviorString(v.(string))
		if error == nil {
			s.LoBehavior = mb
		} else {
			panic(error)
		}
	}
	if v, ok := data["MidBehavior"]; ok {
		mb, error := MaterialBehaviorString(v.(string))
		if error == nil {
			s.MidBehavior = mb
		} else {
			panic(error)
		}
	}
	if v, ok := data["HiBehavior"]; ok {
		mb, error := MaterialBehaviorString(v.(string))
		if error == nil {
			s.HiBehavior = mb
		} else {
			panic(error)
		}
	}
}

func (s *Segment) Serialize() map[string]interface{} {
	result := s.Base.Serialize()
	result["X"] = s.P[0]
	result["Y"] = s.P[1]

	if s.HiMaterial != nil {
		result["HiMaterial"] = s.HiMaterial.GetBase().ID
	}
	if s.LoMaterial != nil {
		result["LoMaterial"] = s.LoMaterial.GetBase().ID
	}
	if s.MidMaterial != nil {
		result["MidMaterial"] = s.MidMaterial.GetBase().ID
	}

	if s.HiBehavior != ScaleNone {
		result["HiBehavior"] = s.HiBehavior.String()
	}
	if s.LoBehavior != ScaleNone {
		result["LoBehavior"] = s.LoBehavior.String()
	}
	if s.MidBehavior != ScaleNone {
		result["MidBehavior"] = s.MidBehavior.String()
	}

	if s.AdjacentSector != nil {
		result["AdjacentSector"] = s.AdjacentSector.GetBase().ID
	}
	return result
}
