package core

import (
	"log"
	"math"

	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
)

type Sector struct {
	concepts.Attached `editable:"^"`

	Segments      []*Segment
	Bodies        map[uint64]*concepts.EntityRef
	BottomZ       concepts.SimScalar  `editable:"Floor Height"`
	TopZ          concepts.SimScalar  `editable:"Ceiling Height"`
	FloorScale    float64             `editable:"Floor Material Scale"`
	CeilScale     float64             `editable:"Ceiling Material Scale"`
	FloorSlope    float64             `editable:"Floor Slope"`
	CeilSlope     float64             `editable:"Ceiling Slope"`
	FloorTarget   *concepts.EntityRef `editable:"Floor Target"`
	CeilTarget    *concepts.EntityRef `editable:"Ceiling Target"`
	FloorMaterial *concepts.EntityRef `editable:"Floor Material" edit_type:"Material"`
	CeilMaterial  *concepts.EntityRef `editable:"Ceiling Material" edit_type:"Material"`
	FloorFriction float64             `editable:"Floor Friction"`
	FloorTriggers []Trigger           `editable:"Floor Triggers"`
	CeilTriggers  []Trigger           `editable:"Ceil Triggers"`

	Winding                           int8
	Min, Max, Center                  concepts.Vector3
	FloorNormal, CeilNormal           concepts.Vector3
	LightmapWidth, LightmapHeight     uint32
	FloorLightmap, CeilLightmap       []concepts.Vector3
	FloorLightmapAge, CeilLightmapAge []int
	PVS                               map[uint64]*Sector
	PVSBody                           map[uint64]*Sector
	PVL                               map[uint64]*concepts.EntityRef

	// RoomImpulse
}

var SectorComponentIndex int

func init() {
	SectorComponentIndex = concepts.DbTypes().Register(Sector{}, SectorFromDb)
}

func SectorFromDb(entity *concepts.EntityRef) *Sector {
	if asserted, ok := entity.Component(SectorComponentIndex).(*Sector); ok {
		return asserted
	}
	return nil
}

func (s *Sector) String() string {
	return "Sector: " + s.Center.StringHuman()
}

func (ms *Sector) IsPointInside2D(p *concepts.Vector2) bool {
	inside := false
	flag1 := (p[1] >= ms.Segments[0].P[1])

	for _, segment := range ms.Segments {
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

func (s *Sector) AddSegment(x float64, y float64) *Segment {
	segment := new(Segment)
	segment.Construct(nil)
	segment.Sector = s
	segment.P = concepts.Vector2{x, y}
	s.Segments = append(s.Segments, segment)
	return segment
}

func (s *Sector) Construct(data map[string]any) {
	s.Attached.Construct(data)
	s.Segments = make([]*Segment, 0)
	s.Bodies = make(map[uint64]*concepts.EntityRef)
	s.BottomZ.Set(0.0)
	s.TopZ.Set(64.0)
	s.FloorScale = 64.0
	s.CeilScale = 64.0
	s.FloorFriction = 0.85
	if data == nil {
		return
	}

	if v, ok := data["TopZ"]; ok {
		s.TopZ.Original = v.(float64)
		s.TopZ.Reset()
	}
	if v, ok := data["BottomZ"]; ok {
		s.BottomZ.Original = v.(float64)
		s.BottomZ.Reset()
	}
	if v, ok := data["FloorScale"]; ok {
		s.FloorScale = v.(float64)
	}
	if v, ok := data["CeilScale"]; ok {
		s.CeilScale = v.(float64)
	}
	if v, ok := data["FloorSlope"]; ok {
		s.FloorSlope = v.(float64)
	}
	if v, ok := data["CeilSlope"]; ok {
		s.CeilSlope = v.(float64)
	}

	if v, ok := data["FloorMaterial"]; ok {
		s.FloorMaterial = s.DB.DeserializeEntityRef(v)
	}
	if v, ok := data["CeilMaterial"]; ok {
		s.CeilMaterial = s.DB.DeserializeEntityRef(v)
	}
	if v, ok := data["Segments"]; ok {
		jsonSegments := v.([]any)
		s.Segments = make([]*Segment, len(jsonSegments))
		for i, jsonSegment := range jsonSegments {
			segment := new(Segment)
			segment.Sector = s
			segment.DB = s.DB
			segment.Construct(jsonSegment.(map[string]any))
			s.Segments[i] = segment
		}
	}
	if v, ok := data["Bodies"]; ok {
		if ers, ok := v.([]any); ok {
			s.Bodies = s.DB.DeserializeEntityRefs(ers)
		}
	}
	if v, ok := data["FloorTarget"]; ok {
		s.FloorTarget = s.DB.DeserializeEntityRef(v)
	}
	if v, ok := data["CeilTarget"]; ok {
		s.CeilTarget = s.DB.DeserializeEntityRef(v)
	}
	if v, ok := data["FloorFriction"]; ok {
		s.FloorFriction = v.(float64)
	}
	if v, ok := data["FloorTriggers"]; ok {
		s.FloorTriggers = ConstructTriggers(v)
	}
	if v, ok := data["CeilTriggers"]; ok {
		s.CeilTriggers = ConstructTriggers(v)
	}

	s.Recalculate()
}

func (s *Sector) Serialize() map[string]any {
	result := s.Attached.Serialize()
	result["TopZ"] = s.TopZ.Original
	result["BottomZ"] = s.BottomZ.Original
	result["FloorScale"] = s.FloorScale
	result["CeilScale"] = s.CeilScale
	result["FloorFriction"] = s.FloorFriction
	if s.FloorSlope != 0 {
		result["FloorSlope"] = s.FloorSlope
	}
	if s.CeilSlope != 0 {
		result["CeilSlope"] = s.CeilSlope
	}

	if !s.FloorTarget.Nil() {
		result["FloorTarget"] = s.FloorTarget.Serialize()
	}
	if !s.CeilTarget.Nil() {
		result["CeilTarget"] = s.CeilTarget.Serialize()
	}
	if !s.FloorMaterial.Nil() {
		result["FloorMaterial"] = s.FloorMaterial.Serialize()
	}
	if !s.CeilMaterial.Nil() {
		result["CeilMaterial"] = s.CeilMaterial.Serialize()
	}
	if len(s.FloorTriggers) > 0 {
		result["FloorTriggers"] = SerializeTriggers(s.FloorTriggers)
	}
	if len(s.CeilTriggers) > 0 {
		result["CeilTriggers"] = SerializeTriggers(s.CeilTriggers)
	}

	if len(s.Bodies) > 0 {
		result["Bodies"] = concepts.SerializeEntityRefMap(s.Bodies)
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
	}

	if sum < 0 {
		s.Winding = 1
	} else {
		s.Winding = -1
	}

	filtered := make([]*Segment, 0)
	var prev *Segment
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

		if s.Winding < 0 {
			segment.Normal.MulSelf(-1)
		}
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

	s.LightmapWidth = uint32((s.Max[0]-s.Min[0])/constants.LightGrid) + constants.LightSafety*2
	s.LightmapHeight = uint32((s.Max[1]-s.Min[1])/constants.LightGrid) + constants.LightSafety*2
	s.FloorLightmap = make([]concepts.Vector3, s.LightmapWidth*s.LightmapHeight)
	s.CeilLightmap = make([]concepts.Vector3, s.LightmapWidth*s.LightmapHeight)
	s.FloorLightmapAge = make([]int, s.LightmapWidth*s.LightmapHeight)
	s.CeilLightmapAge = make([]int, s.LightmapWidth*s.LightmapHeight)
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
	return s.slopedZ(s.BottomZ.Render, s.TopZ.Render, isect)
}

// SlopedZOriginal figures out the current slice Z values accounting for slope.
func (s *Sector) SlopedZOriginal(isect *concepts.Vector2) (floorZ float64, ceilZ float64) {
	return s.slopedZ(s.BottomZ.Original, s.TopZ.Original, isect)
}

// SlopedZNow figures out the current slice Z values accounting for slope.
func (s *Sector) SlopedZNow(isect *concepts.Vector2) (floorZ float64, ceilZ float64) {
	return s.slopedZ(s.BottomZ.Now, s.TopZ.Now, isect)
}

func (s *Sector) LightmapAddress(p *concepts.Vector2) uint32 {
	dx := int((p[0]-s.Min[0])/constants.LightGrid) + constants.LightSafety
	dy := int((p[1]-s.Min[1])/constants.LightGrid) + constants.LightSafety
	dx = concepts.IntClamp(dx, 0, int(s.LightmapWidth-constants.LightSafety))  // Leave column safety for bi-linear filtering
	dy = concepts.IntClamp(dy, 0, int(s.LightmapHeight-constants.LightSafety)) // Leave row safety for bi-linear filtering
	return uint32(dy)*s.LightmapWidth + uint32(dx)
}

func (s *Sector) ToLightmapWorld(p *concepts.Vector3, floor bool) *concepts.Vector3 {
	lw := p
	lw.SubSelf(&s.Min).MulSelf(1.0 / constants.LightGrid)
	lw[0] = math.Floor(lw[0])*constants.LightGrid + s.Min[0]
	lw[1] = math.Floor(lw[1])*constants.LightGrid + s.Min[1]
	floorZ, ceilZ := s.SlopedZRender(lw.To2D())
	if floor {
		lw[2] = floorZ
	} else {
		lw[2] = ceilZ
	}
	return lw
}

func (s *Sector) LightmapAddressToWorld(r *concepts.Vector3, mapIndex uint32, floor bool) *concepts.Vector3 {
	u := int(mapIndex%s.LightmapWidth) - constants.LightSafety
	v := int(mapIndex/s.LightmapWidth) - constants.LightSafety
	r[0] = s.Min[0] + (float64(u)+0.0)*constants.LightGrid
	r[1] = s.Min[1] + (float64(v)+0.0)*constants.LightGrid
	floorZ, ceilZ := s.SlopedZRender(r.To2D())
	if floor {
		r[2] = floorZ
	} else {
		r[2] = ceilZ
	}
	return r
}

func (s *Sector) InBounds(world *concepts.Vector3) bool {
	return (world[0] >= s.Min[0]-constants.IntersectEpsilon && world[0] <= s.Max[0]+constants.IntersectEpsilon &&
		world[1] >= s.Min[1]-constants.IntersectEpsilon && world[1] <= s.Max[1]+constants.IntersectEpsilon &&
		world[2] >= s.Min[2]-constants.IntersectEpsilon && world[2] <= s.Max[2]+constants.IntersectEpsilon)
}

func (s *Sector) AABBIntersect(min, max *concepts.Vector3) bool {
	dx := (s.Min[0] + s.Max[0] - min[0] - max[0]) * 0.5
	px := (s.Max[0]-s.Min[0]+max[0]-min[0])*0.5 - math.Abs(dx)
	if px <= 0 {
		return false
	}

	dy := (s.Min[1] + s.Max[1] - min[1] - max[1]) * 0.5
	py := (s.Max[1] - s.Min[1] + max[1] - min[1]) - math.Abs(dy)

	return py > 0
}

func (s *Sector) DbgPrintSegments() {
	log.Printf("Segments:")
	for i, seg := range s.Segments {
		log.Printf("%v: %v", i, seg.P.StringHuman())
	}
}
