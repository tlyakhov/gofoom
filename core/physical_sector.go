package core

import (
	"math"
	"reflect"

	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/registry"
)

type PhysicalSector struct {
	concepts.Base `editable:"^"`

	Map           *Map
	Segments      []*Segment
	Entities      map[string]AbstractEntity
	BottomZ       float64        `editable:"Floor Height"`
	TopZ          float64        `editable:"Ceiling Height"`
	FloorScale    float64        `editable:"Floor Material Scale"`
	CeilScale     float64        `editable:"Ceiling Material Scale"`
	FloorSlope    float64        `editable:"Floor Slope"`
	CeilSlope     float64        `editable:"Ceiling Slope"`
	FloorTarget   AbstractSector `editable:"Floor Target"`
	CeilTarget    AbstractSector `editable:"Ceiling Target"`
	FloorMaterial Sampleable     `editable:"Floor Material" edit_type:"Material"`
	CeilMaterial  Sampleable     `editable:"Ceiling Material" edit_type:"Material"`

	Winding                           int8
	Min, Max, Center                  concepts.Vector3
	FloorNormal, CeilNormal           concepts.Vector3
	LightmapWidth, LightmapHeight     uint32
	FloorLightmap, CeilLightmap       []concepts.Vector3
	FloorLightmapAge, CeilLightmapAge []int
	PVS                               map[string]AbstractSector
	PVSEntity                         map[string]AbstractSector
	PVL                               map[string]AbstractEntity

	// RoomImpulse
}

func init() {
	registry.Instance().Register(PhysicalSector{})
}

func (s *PhysicalSector) Physical() *PhysicalSector {
	return s
}

func (ms *PhysicalSector) IsPointInside2D(p *concepts.Vector2) bool {
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

func (s *PhysicalSector) Initialize() {
	s.Base.Initialize()
	s.Segments = make([]*Segment, 0)
	s.Entities = make(map[string]AbstractEntity)
	s.BottomZ = 0.0
	s.TopZ = 64.0
	s.FloorScale = 64.0
	s.CeilScale = 64.0
}

func (s *PhysicalSector) AddSegment(x float64, y float64) *Segment {
	segment := &Segment{}
	segment.Initialize()
	segment.SetParent(s)
	segment.P = concepts.Vector2{x, y}
	s.Segments = append(s.Segments, segment)
	return segment
}

func (s *PhysicalSector) Deserialize(data map[string]interface{}) {
	s.Initialize()
	s.Base.Deserialize(data)
	if v, ok := data["TopZ"]; ok {
		s.TopZ = v.(float64)
	}
	if v, ok := data["BottomZ"]; ok {
		s.BottomZ = v.(float64)
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
		s.FloorMaterial = s.Map.Materials[v.(string)]
	}
	if v, ok := data["CeilMaterial"]; ok {
		s.CeilMaterial = s.Map.Materials[v.(string)]
	}
	if v, ok := data["Segments"]; ok {
		concepts.MapArray(s, &s.Segments, v)
	}
	if v, ok := data["Entities"]; ok {
		concepts.MapCollection(s, &s.Entities, v)
	}
	if v, ok := data["FloorTarget"]; ok {
		s.FloorTarget = &PlaceholderSector{Base: concepts.Base{ID: v.(string)}}
	}
	if v, ok := data["CeilTarget"]; ok {
		s.CeilTarget = &PlaceholderSector{Base: concepts.Base{ID: v.(string)}}
	}
	s.Recalculate()
}

func (s *PhysicalSector) Serialize() map[string]interface{} {
	result := s.Base.Serialize()
	result["Type"] = "core.PhysicalSector"
	result["TopZ"] = s.TopZ
	result["BottomZ"] = s.BottomZ
	result["FloorScale"] = s.FloorScale
	result["CeilScale"] = s.CeilScale
	if s.FloorSlope != 0 {
		result["FloorSlope"] = s.FloorSlope
	}
	if s.CeilSlope != 0 {
		result["CeilSlope"] = s.CeilSlope
	}

	if s.FloorTarget != nil {
		result["FloorTarget"] = s.FloorTarget.GetBase().ID
	}
	if s.CeilTarget != nil {
		result["CeilTarget"] = s.CeilTarget.GetBase().ID
	}
	if s.FloorMaterial != nil {
		result["FloorMaterial"] = s.FloorMaterial.GetBase().ID
	}
	if s.CeilMaterial != nil {
		result["CeilMaterial"] = s.CeilMaterial.GetBase().ID
	}

	entities := []interface{}{}
	for _, e := range s.Entities {
		// Don't serialize player
		if reflect.TypeOf(e).String() == "*entities.Player" {
			continue
		}
		entities = append(entities, e.Serialize())
	}
	result["Entities"] = entities

	segments := []interface{}{}
	for _, seg := range s.Segments {
		segments = append(segments, seg.Serialize())
	}
	result["Segments"] = segments
	return result
}

func (s *PhysicalSector) Recalculate() {
	concepts.V3(&s.Center, 0, 0, (s.TopZ+s.BottomZ)/2)
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

	filtered := s.Segments[:0]
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
		floorZ, ceilZ := s.CalcFloorCeilingZ(&segment.P)
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
		sloped := s.Segments[0].Normal.To3D(&concepts.Vector3{})
		delta := (&concepts.Vector2{s.Segments[1].P[0], s.Segments[1].P[1]}).Sub(&s.Segments[0].P).To3D(&concepts.Vector3{})
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

	if ph, ok := s.FloorTarget.(*PlaceholderSector); ok {
		// Get the actual one.
		if actual, ok := s.Map.Sectors[ph.ID]; ok {
			s.FloorTarget = actual
		}
	}

	if ph, ok := s.CeilTarget.(*PlaceholderSector); ok {
		// Get the actual one.
		if actual, ok := s.Map.Sectors[ph.ID]; ok {
			s.CeilTarget = actual
		}
	}

	s.LightmapWidth = uint32((s.Max[0]-s.Min[0])/constants.LightGrid) + constants.LightSafety*2
	s.LightmapHeight = uint32((s.Max[1]-s.Min[1])/constants.LightGrid) + constants.LightSafety*2
	s.FloorLightmap = make([]concepts.Vector3, s.LightmapWidth*s.LightmapHeight)
	s.CeilLightmap = make([]concepts.Vector3, s.LightmapWidth*s.LightmapHeight)
	s.FloorLightmapAge = make([]int, s.LightmapWidth*s.LightmapHeight)
	s.CeilLightmapAge = make([]int, s.LightmapWidth*s.LightmapHeight)
}

// CalcFloorCeilingZ figures out the current slice Z values accounting for slope.
func (s *PhysicalSector) CalcFloorCeilingZ(isect *concepts.Vector2) (floorZ float64, ceilZ float64) {
	if len(s.Segments) < 2 {
		return s.BottomZ, s.TopZ
	}
	dist := 0.0
	if s.FloorSlope != 0 || s.CeilSlope != 0 {
		a := &s.Segments[0].P
		b := &s.Segments[1].P
		length2 := a.Dist2(b)
		if length2 == 0 {
			dist = isect.Dist(a)
		} else {
			delta := concepts.Vector2{b[0] - a[0], b[1] - a[1]}
			t := ((isect[0]-a[0])*delta[0] + (isect[1]-a[1])*delta[1]) / length2
			delta[0] = a[0] + delta[0]*t
			delta[1] = a[1] + delta[1]*t
			dist = isect.Dist(&delta)
		}
	}

	if s.FloorSlope == 0 {
		floorZ = s.BottomZ
	} else {
		floorZ = s.BottomZ + s.FloorSlope*dist
	}

	if s.CeilSlope == 0 {
		ceilZ = s.TopZ
	} else {
		ceilZ = s.TopZ + s.CeilSlope*dist
	}
	return
}

func (s *PhysicalSector) LightmapAddress(p *concepts.Vector2) uint32 {
	dx := int((p[0]-s.Min[0])/constants.LightGrid) + constants.LightSafety
	dy := int((p[1]-s.Min[1])/constants.LightGrid) + constants.LightSafety
	dx = concepts.IntClamp(dx, 0, int(s.LightmapWidth))
	dy = concepts.IntClamp(dy, 0, int(s.LightmapHeight))
	return uint32(dy)*s.LightmapWidth + uint32(dx)
}

func (s *PhysicalSector) ToLightmapWorld(p *concepts.Vector3, floor bool) *concepts.Vector3 {
	lw := p
	lw.SubSelf(&s.Min).MulSelf(1.0 / constants.LightGrid)
	lw[0] = math.Floor(lw[0])*constants.LightGrid + s.Min[0]
	lw[1] = math.Floor(lw[1])*constants.LightGrid + s.Min[1]
	floorZ, ceilZ := s.CalcFloorCeilingZ(lw.To2D())
	if floor {
		lw[2] = floorZ
	} else {
		lw[2] = ceilZ
	}
	return lw
}

func (s *PhysicalSector) LightmapAddressToWorld(r *concepts.Vector3, mapIndex uint32, floor bool) *concepts.Vector3 {
	u := int(mapIndex%s.LightmapWidth) - constants.LightSafety
	v := int(mapIndex/s.LightmapWidth) - constants.LightSafety
	r[0] = s.Min[0] + (float64(u)+0.0)*constants.LightGrid
	r[1] = s.Min[1] + (float64(v)+0.0)*constants.LightGrid
	floorZ, ceilZ := s.CalcFloorCeilingZ(r.To2D())
	if floor {
		r[2] = floorZ
	} else {
		r[2] = ceilZ
	}
	return r
}

func (s *PhysicalSector) SetParent(parent interface{}) {
	if m, ok := parent.(*Map); ok {
		s.Map = m
	} else {
		panic("Tried core.PhysicalSector.SetParent with a parameter that wasn't a *core.Map.")
	}
}
