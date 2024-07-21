// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package core

import (
	"log"
	"math"

	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"

	"github.com/puzpuzpuz/xsync/v3"
)

type Sector struct {
	concepts.Attached `editable:"^"`

	Segments         []*SectorSegment
	Bodies           map[concepts.Entity]*Body
	InternalSegments map[concepts.Entity]*InternalSegment
	BottomZ          concepts.SimVariable[float64] `editable:"Floor Height"`
	TopZ             concepts.SimVariable[float64] `editable:"Ceiling Height"`
	FloorSlope       float64                       `editable:"Floor Slope"`
	CeilSlope        float64                       `editable:"Ceiling Slope"`
	FloorTarget      concepts.Entity               `editable:"Floor Target" edit_type:"Sector"`
	CeilTarget       concepts.Entity               `editable:"Ceiling Target" edit_type:"Sector"`
	FloorSurface     materials.Surface             `editable:"Floor Surface"`
	CeilSurface      materials.Surface             `editable:"Ceiling Surface"`
	Gravity          concepts.Vector3              `editable:"Gravity"`
	FloorFriction    float64                       `editable:"Floor Friction"`
	FloorScripts     []*Script                     `editable:"Floor Scripts"`
	CeilScripts      []*Script                     `editable:"Ceil Scripts"`
	EnterScripts     []*Script                     `editable:"Enter Scripts"`
	ExitScripts      []*Script                     `editable:"Exit Scripts"`

	Winding                 int8
	Min, Max, Center        concepts.Vector3
	FloorNormal, CeilNormal concepts.Vector3
	PVS                     map[concepts.Entity]*Sector
	PVL                     map[concepts.Entity]*Body
	Lightmap                *xsync.MapOf[uint64, concepts.Vector4]
	LightmapBias            [3]int64 // Quantized Min
}

var SectorComponentIndex int

func init() {
	SectorComponentIndex = concepts.DbTypes().Register(Sector{}, SectorFromDb)
}

func SectorFromDb(db *concepts.EntityComponentDB, e concepts.Entity) *Sector {
	if asserted, ok := db.Component(e, SectorComponentIndex).(*Sector); ok {
		return asserted
	}
	return nil
}

func (s *Sector) String() string {
	return "Sector: " + s.Center.StringHuman()
}

func (s *Sector) IsPointInside2D(p *concepts.Vector2) bool {
	inside := false
	flag1 := (p[1] >= s.Segments[0].P[1])

	for _, segment := range s.Segments {
		flag2 := (p[1] >= segment.Next.P[1])
		if flag1 != flag2 {
			if ((segment.Next.P[1]-p[1])*(segment.P[0]-segment.Next.P[0]) >= (segment.Next.P[0]-p[0])*(segment.P[1]-segment.Next.P[1])) == flag2 {
				inside = !inside
			}
		}
		flag1 = flag2
	}
	return inside
}

func (s *Sector) SetDB(db *concepts.EntityComponentDB) {
	if s.DB != nil {
		s.TopZ.Detach(s.DB.Simulation)
		s.BottomZ.Detach(s.DB.Simulation)
	}
	s.Attached.SetDB(db)
	s.TopZ.Attach(db.Simulation)
	s.BottomZ.Attach(db.Simulation)
}

func (s *Sector) AddSegment(x float64, y float64) *SectorSegment {
	segment := new(SectorSegment)
	segment.Construct(s.DB, nil)
	segment.Sector = s
	segment.P = concepts.Vector2{x, y}
	s.Segments = append(s.Segments, segment)
	return segment
}

var defaultSectorTopZ = map[string]any{"Original": 64.0}

func (s *Sector) Construct(data map[string]any) {
	s.Attached.Construct(data)
	s.Lightmap = xsync.NewMapOf[uint64, concepts.Vector4]()
	s.Segments = make([]*SectorSegment, 0)
	s.Bodies = make(map[concepts.Entity]*Body)
	s.InternalSegments = make(map[concepts.Entity]*InternalSegment)
	s.Gravity[0] = 0
	s.Gravity[1] = 0
	s.Gravity[2] = -constants.Gravity
	s.BottomZ.Construct(nil)
	s.TopZ.Construct(defaultSectorTopZ)
	s.FloorFriction = 0.85
	s.FloorSurface.Construct(s.DB, nil)
	s.CeilSurface.Construct(s.DB, nil)

	if data == nil {
		return
	}

	if v, ok := data["TopZ"]; ok {
		if v2, ok2 := v.(float64); ok2 {
			v = map[string]any{"Original": v2}
		}
		s.TopZ.Construct(v.(map[string]any))
	}
	if v, ok := data["BottomZ"]; ok {
		if v2, ok2 := v.(float64); ok2 {
			v = map[string]any{"Original": v2}
		}
		s.BottomZ.Construct(v.(map[string]any))
	}
	if v, ok := data["FloorSlope"]; ok {
		s.FloorSlope = v.(float64)
	}
	if v, ok := data["CeilSlope"]; ok {
		s.CeilSlope = v.(float64)
	}

	if v, ok := data["FloorSurface"]; ok {
		s.FloorSurface.Construct(s.DB, v.(map[string]any))
	}
	if v, ok := data["CeilSurface"]; ok {
		s.CeilSurface.Construct(s.DB, v.(map[string]any))
	}
	if v, ok := data["Segments"]; ok {
		jsonSegments := v.([]any)
		s.Segments = make([]*SectorSegment, len(jsonSegments))
		for i, jsonSegment := range jsonSegments {
			segment := new(SectorSegment)
			segment.Sector = s
			segment.Construct(s.DB, jsonSegment.(map[string]any))
			s.Segments[i] = segment
		}
	}
	if v, ok := data["FloorTarget"]; ok {
		s.FloorTarget, _ = concepts.ParseEntity(v.(string))
	}
	if v, ok := data["CeilTarget"]; ok {
		s.CeilTarget, _ = concepts.ParseEntity(v.(string))
	}
	if v, ok := data["Gravity"]; ok {
		s.Gravity.Deserialize(v.(map[string]any))
	}
	if v, ok := data["FloorFriction"]; ok {
		s.FloorFriction = v.(float64)
	}
	if v, ok := data["FloorScripts"]; ok {
		s.FloorScripts = concepts.ConstructSlice[*Script](s.DB, v)
	}
	if v, ok := data["CeilScripts"]; ok {
		s.CeilScripts = concepts.ConstructSlice[*Script](s.DB, v)
	}
	if v, ok := data["EnterScripts"]; ok {
		s.EnterScripts = concepts.ConstructSlice[*Script](s.DB, v)
	}
	if v, ok := data["ExitScripts"]; ok {
		s.ExitScripts = concepts.ConstructSlice[*Script](s.DB, v)
	}

	s.Recalculate()
}

func (s *Sector) Serialize() map[string]any {
	result := s.Attached.Serialize()
	result["TopZ"] = s.TopZ.Serialize()
	result["BottomZ"] = s.BottomZ.Serialize()
	result["FloorSurface"] = s.FloorSurface.Serialize()
	result["CeilSurface"] = s.CeilSurface.Serialize()

	if s.Gravity[0] != 0 || s.Gravity[1] != 0 || s.Gravity[2] != -constants.Gravity {
		result["Gravity"] = s.Gravity.Serialize()
	}
	result["FloorFriction"] = s.FloorFriction
	if s.FloorSlope != 0 {
		result["FloorSlope"] = s.FloorSlope
	}
	if s.CeilSlope != 0 {
		result["CeilSlope"] = s.CeilSlope
	}

	if s.FloorTarget != 0 {
		result["FloorTarget"] = s.FloorTarget.Format()
	}
	if s.CeilTarget != 0 {
		result["CeilTarget"] = s.CeilTarget.Format()
	}
	if len(s.FloorScripts) > 0 {
		result["FloorScripts"] = concepts.SerializeSlice(s.FloorScripts)
	}
	if len(s.CeilScripts) > 0 {
		result["CeilScripts"] = concepts.SerializeSlice(s.CeilScripts)
	}
	if len(s.EnterScripts) > 0 {
		result["EnterScripts"] = concepts.SerializeSlice(s.EnterScripts)
	}
	if len(s.ExitScripts) > 0 {
		result["ExitScripts"] = concepts.SerializeSlice(s.ExitScripts)
	}

	segments := []any{}
	for _, seg := range s.Segments {
		segments = append(segments, seg.Serialize())
	}
	result["Segments"] = segments
	return result
}

func (s *Sector) Recalculate() {
	concepts.V3(&s.Center, 0, 0, (s.TopZ.Original+s.BottomZ.Original)/2)
	concepts.V3(&s.Min, math.Inf(1), math.Inf(1), math.Inf(1))
	concepts.V3(&s.Max, math.Inf(-1), math.Inf(-1), math.Inf(-1))

	sum := 0.0
	for i, segment := range s.Segments {
		// Can't use prev/next pointers because they haven't been initialized yet.
		next := s.Segments[(i+1)%len(s.Segments)]
		sum += (next.P[0] - segment.P[0]) * (segment.P[1] + next.P[1])
		segment.Index = i
	}

	if sum < 0 {
		s.Winding = 1
	} else {
		s.Winding = -1
	}

	filtered := make([]*SectorSegment, 0)
	var prev *SectorSegment
	for i, segment := range s.Segments {
		next := s.Segments[(i+1)%len(s.Segments)]
		// Filter out degenerate segments.
		if prev != nil && prev.P == segment.P {
			prev.Next = next
			next.Prev = prev
			continue
		}
		filtered = append(filtered, segment)
		segment.Next = next
		next.Prev = segment
		prev = segment
		s.Center[0] += segment.P[0]
		s.Center[1] += segment.P[1]
		if segment.P[0] < s.Min[0] {
			s.Min[0] = segment.P[0]
		}
		if segment.P[1] < s.Min[1] {
			s.Min[1] = segment.P[1]
		}
		if segment.P[0] > s.Max[0] {
			s.Max[0] = segment.P[0]
		}
		if segment.P[1] > s.Max[1] {
			s.Max[1] = segment.P[1]
		}
		floorZ, ceilZ := s.SlopedZOriginal(&segment.P)
		if floorZ < s.Min[2] {
			s.Min[2] = floorZ
		}
		if ceilZ < s.Min[2] {
			s.Min[2] = ceilZ
		}
		if floorZ > s.Max[2] {
			s.Max[2] = floorZ
		}
		if ceilZ > s.Max[2] {
			s.Max[2] = ceilZ
		}
		segment.Sector = s
		segment.Recalculate()
	}
	s.Segments = filtered

	if len(s.Segments) > 1 {
		sloped := s.Segments[0].Normal.To3D(new(concepts.Vector3))
		delta := (&concepts.Vector2{s.Segments[1].P[0], s.Segments[1].P[1]}).Sub(&s.Segments[0].P).To3D(new(concepts.Vector3))
		sloped[2] = s.FloorSlope
		s.FloorNormal.CrossSelf(sloped, delta).NormSelf()
		if s.FloorNormal[2] < 0 {
			s.FloorNormal.MulSelf(-1)
		}
		s.Segments[0].Normal.To3D(sloped)
		sloped[2] = s.CeilSlope
		s.CeilNormal.CrossSelf(sloped, delta).NormSelf()
		if s.CeilNormal[2] > 0 {
			s.CeilNormal.MulSelf(-1)
		}
	}

	s.Center.MulSelf(1.0 / float64(len(s.Segments)))
	// Floor is important, needs to truncate towards -Infinity rather than 0
	s.LightmapBias[0] = int64(math.Floor(s.Min[0]/constants.LightGrid)) - 20
	s.LightmapBias[1] = int64(math.Floor(s.Min[1]/constants.LightGrid)) - 20
	s.LightmapBias[2] = int64(math.Floor(s.Min[2]/constants.LightGrid)) - 20
}

const lightmapMask uint64 = (1 << 16) - 1

func (s *Sector) WorldToLightmapAddress(v *concepts.Vector3, flags uint16) uint64 {
	// Floor is important, needs to truncate towards -Infinity rather than 0
	z := int64(math.Floor(v[2]/constants.LightGrid)) - s.LightmapBias[2]
	y := int64(math.Floor(v[1]/constants.LightGrid)) - s.LightmapBias[1]
	x := int64(math.Floor(v[0]/constants.LightGrid)) - s.LightmapBias[0]
	/*if x < 0 || y < 0 || z < 0 {
		fmt.Printf("Error: lightmap address conversion resulted in negative value: %v,%v,%v\n", x, y, z)
	}*/
	// Bit shift and mask the components, and add the sector entity at the end
	// to ensure that overlapping addresses are distinct for each sector
	return (((uint64(x) & lightmapMask) << 48) |
		((uint64(y) & lightmapMask) << 32) |
		((uint64(z) & lightmapMask) << 16) |
		uint64(flags)) //+ (s.Entity * 1009)
}

func (s *Sector) LightmapAddressToWorld(result *concepts.Vector3, a uint64) *concepts.Vector3 {
	//w := uint64(a & wMask)
	//a -= s.Entity * 1009
	a = a >> 16
	z := int64((a & lightmapMask)) + s.LightmapBias[2]
	result[2] = float64(z) * constants.LightGrid
	a = a >> 16
	y := int64((a & lightmapMask)) + s.LightmapBias[1]
	result[1] = float64(y) * constants.LightGrid
	a = a >> 16
	x := int64((a & lightmapMask)) + s.LightmapBias[0]
	result[0] = float64(x) * constants.LightGrid
	return result
}

func (s *Sector) CalcDistforZ(isect *concepts.Vector2) float64 {
	a := &s.Segments[0].P
	b := &s.Segments[1].P
	length2 := a.Dist2(b)
	if length2 == 0 {
		return isect.Dist(a)
	} else {
		delta := concepts.Vector2{b[0] - a[0], b[1] - a[1]}
		t := ((isect[0]-a[0])*delta[0] + (isect[1]-a[1])*delta[1]) / length2
		delta[0] = a[0] + delta[0]*t
		delta[1] = a[1] + delta[1]*t
		return isect.Dist(&delta)
	}
}

func (s *Sector) slopedZ(fz, cz float64, isect *concepts.Vector2) (floorZ float64, ceilZ float64) {
	floorZ = fz
	ceilZ = cz
	// Fast path
	if (s.FloorSlope == 0 && s.CeilSlope == 0) || len(s.Segments) < 2 {
		return
	}

	dist := s.CalcDistforZ(isect)

	if s.FloorSlope != 0 {
		floorZ += s.FloorSlope * dist
	}
	if s.CeilSlope != 0 {
		ceilZ += s.CeilSlope * dist
	}
	return
}

// SlopedZRender figures out the current slice Z values accounting for slope.
func (s *Sector) SlopedZRender(isect *concepts.Vector2) (floorZ float64, ceilZ float64) {
	return s.slopedZ(*s.BottomZ.Render, *s.TopZ.Render, isect)
}

// SlopedZOriginal figures out the current slice Z values accounting for slope.
func (s *Sector) SlopedZOriginal(isect *concepts.Vector2) (floorZ float64, ceilZ float64) {
	return s.slopedZ(s.BottomZ.Original, s.TopZ.Original, isect)
}

// SlopedZNow figures out the current slice Z values accounting for slope.
func (s *Sector) SlopedZNow(isect *concepts.Vector2) (floorZ float64, ceilZ float64) {
	return s.slopedZ(s.BottomZ.Now, s.TopZ.Now, isect)
}

func (s *Sector) InBounds(world *concepts.Vector3) bool {
	return (world[0] >= s.Min[0]-constants.IntersectEpsilon && world[0] <= s.Max[0]+constants.IntersectEpsilon &&
		world[1] >= s.Min[1]-constants.IntersectEpsilon && world[1] <= s.Max[1]+constants.IntersectEpsilon &&
		world[2] >= s.Min[2]-constants.IntersectEpsilon && world[2] <= s.Max[2]+constants.IntersectEpsilon)
}

func (s *Sector) AABBIntersect(min, max *concepts.Vector3, includeEdges bool) bool {
	if includeEdges {
		return (s.Min[0] <= max[0] &&
			s.Max[0] >= min[0] &&
			s.Min[1] <= max[1] &&
			s.Max[1] >= min[1] &&
			s.Min[2] <= max[2] &&
			s.Max[2] >= min[2])
	} else {
		return (s.Min[0] < max[0] &&
			s.Max[0] > min[0] &&
			s.Min[1] < max[1] &&
			s.Max[1] > min[1] &&
			s.Min[2] < max[2] &&
			s.Max[2] > min[2])
	}
}

func (s *Sector) AABBIntersect2D(min, max *concepts.Vector2, includeEdges bool) bool {
	if includeEdges {
		return (s.Min[0] <= max[0] &&
			s.Max[0] >= min[0] &&
			s.Min[1] <= max[1] &&
			s.Max[1] >= min[1])
	} else {
		return (s.Min[0] < max[0] &&
			s.Max[0] > min[0] &&
			s.Min[1] < max[1] &&
			s.Max[1] > min[1])
	}
}

func (s *Sector) DbgPrintSegments() {
	log.Printf("Segments:")
	for i, seg := range s.Segments {
		log.Printf("%v: %v", i, seg.P.StringHuman())
	}
}
