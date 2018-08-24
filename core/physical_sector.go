package core

import (
	"math"

	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/constants"
	"github.com/tlyakhov/gofoom/registry"
)

type PhysicalSector struct {
	concepts.Base `editable:"^"`

	Map           *Map
	Segments      []*Segment
	Entities      map[string]AbstractEntity
	BottomZ       float64                `editable:"Floor Height"`
	TopZ          float64                `editable:"Ceiling Height"`
	FloorScale    float64                `editable:"Floor Material Scale"`
	CeilScale     float64                `editable:"Ceiling Material Scale"`
	FloorSlope    float64                `editable:"Floor Slope"`
	CeilSlope     float64                `editable:"Ceiling Slope"`
	FloorTarget   AbstractSector         `editable:"Floor Target"`
	CeilTarget    AbstractSector         `editable:"Ceiling Target"`
	FloorMaterial concepts.ISerializable `editable:"Floor Material"`
	CeilMaterial  concepts.ISerializable `editable:"Ceiling Material"`

	Min, Max, Center              concepts.Vector3
	FloorNormal, CeilNormal       concepts.Vector3
	LightmapWidth, LightmapHeight uint32
	FloorLightmap, CeilLightmap   []concepts.Vector3
	PVS                           map[string]AbstractSector
	PVSEntity                     map[string]AbstractSector
	PVSLights                     []AbstractEntity

	// RoomImpulse
}

func init() {
	registry.Instance().Register(PhysicalSector{})
}

func (s *PhysicalSector) Physical() *PhysicalSector {
	return s
}

func (ms *PhysicalSector) IsPointInside2D(p concepts.Vector2) bool {
	inside := false
	flag1 := (p.Y >= ms.Segments[0].A.Y)

	for _, segment := range ms.Segments {
		flag2 := (p.Y >= segment.B.Y)
		if flag1 != flag2 {
			if ((segment.B.Y-p.Y)*(segment.A.X-segment.B.X) >= (segment.B.X-p.X)*(segment.A.Y-segment.B.Y)) == flag2 {
				inside = !inside
			}
		}
		flag1 = flag2
	}
	return inside
}

func (ms *PhysicalSector) Winding() bool {
	sum := 0.0
	for i, segment := range ms.Segments {
		next := ms.Segments[(i+1)%len(ms.Segments)]
		sum += (next.A.X - segment.A.X) * (segment.A.Y + next.A.Y)
	}
	return sum < 0
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
	s.Center = concepts.Vector3{0, 0, (s.TopZ + s.BottomZ) / 2}
	s.Min = concepts.Vector3{math.Inf(1), math.Inf(1), s.BottomZ}
	s.Max = concepts.Vector3{math.Inf(-1), math.Inf(-1), s.TopZ}

	w := s.Winding()

	filtered := s.Segments[:0]
	var prev *Segment
	for i, segment := range s.Segments {
		nextIndex := (i + 1) % len(s.Segments)
		next := s.Segments[nextIndex]
		// Filter out degenerate segments.
		if prev != nil && prev.A == segment.A {
			continue
		}
		filtered = append(filtered, segment)
		prev = segment

		s.Center.X += segment.A.X
		s.Center.Y += segment.A.Y
		if segment.A.X < s.Min.X {
			s.Min.X = segment.A.X
		}
		if segment.A.Y < s.Min.Y {
			s.Min.Y = segment.A.Y
		}
		if segment.A.X > s.Max.X {
			s.Max.X = segment.A.X
		}
		if segment.A.Y > s.Max.Y {
			s.Max.Y = segment.A.Y
		}
		segment.Sector = s
		segment.B = next.A
		segment.Recalculate()

		if !w {
			segment.Normal = segment.Normal.Mul(-1)
		}
	}
	s.Segments = filtered

	if len(s.Segments) > 0 {
		sloped := s.Segments[0].Normal.To3D()
		delta := s.Segments[0].B.Sub(s.Segments[0].A).To3D()
		sloped.Z = s.FloorSlope
		s.FloorNormal = sloped.Cross(delta).Norm()
		if s.FloorNormal.Z < 0 {
			s.FloorNormal.X = -s.FloorNormal.X
			s.FloorNormal.Y = -s.FloorNormal.Y
			s.FloorNormal.Z = -s.FloorNormal.Z
		}
		sloped = s.Segments[0].Normal.To3D()
		sloped.Z = s.CeilSlope
		s.CeilNormal = sloped.Cross(delta).Norm()
		if s.CeilNormal.Z > 0 {
			s.CeilNormal.X = -s.CeilNormal.X
			s.CeilNormal.Y = -s.CeilNormal.Y
			s.CeilNormal.Z = -s.CeilNormal.Z
		}
	}

	s.Center = s.Center.Mul(1.0 / float64(len(s.Segments)))

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

	s.LightmapWidth = uint32((s.Max.X-s.Min.X)/constants.LightGrid) + constants.LightSafety*2
	s.LightmapHeight = uint32((s.Max.Y-s.Min.Y)/constants.LightGrid) + constants.LightSafety*2
	s.FloorLightmap = make([]concepts.Vector3, s.LightmapWidth*s.LightmapHeight)
	s.CeilLightmap = make([]concepts.Vector3, s.LightmapWidth*s.LightmapHeight)
	s.ClearLightmaps()
}

func (s *PhysicalSector) ClearLightmaps() {
	for i := range s.FloorLightmap {
		s.FloorLightmap[i] = concepts.Vector3{-1, -1, -1}
		s.CeilLightmap[i] = concepts.Vector3{-1, -1, -1}
	}

	for _, segment := range s.Segments {
		segment.ClearLightmap()
	}
}

// CalcFloorCeilingZ figures out the current slice Z values accounting for slope.
func (s *PhysicalSector) CalcFloorCeilingZ(isect concepts.Vector2) (floorZ float64, ceilZ float64) {
	dist := 0.0
	if s.FloorSlope != 0 || s.CeilSlope != 0 {
		first := s.Segments[0]
		length2 := first.A.Dist2(first.B)
		if length2 == 0 {
			dist = isect.Dist2(first.A)
		} else {
			delta := first.B.Sub(first.A)
			t := isect.Sub(first.A).Dot(delta) / length2
			dist = isect.Dist(first.A.Add(delta.Mul(t)))
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

func (s *PhysicalSector) LightmapAddress(p concepts.Vector2) uint32 {
	dx := int(p.X-s.Min.X)/constants.LightGrid + constants.LightSafety
	dy := int(p.Y-s.Min.Y)/constants.LightGrid + constants.LightSafety
	dx = concepts.IntClamp(dx, 0, int(s.LightmapWidth))
	dy = concepts.IntClamp(dy, 0, int(s.LightmapHeight))
	return uint32(dy)*s.LightmapWidth + uint32(dx)
}

func (s *PhysicalSector) LightmapWorld(p concepts.Vector3) concepts.Vector3 {
	lw := p.Sub(s.Min).Mul(1.0 / constants.LightGrid)
	lw.X = math.Floor(lw.X) * constants.LightGrid
	lw.Y = math.Floor(lw.Y) * constants.LightGrid
	lw.Z = p.Z
	return lw
}

func (s *PhysicalSector) LightmapAddressToWorld(mapIndex uint32, floor bool) concepts.Vector3 {
	u := int(mapIndex%s.LightmapWidth) - constants.LightSafety
	v := int(mapIndex/s.LightmapWidth) - constants.LightSafety
	r := concepts.Vector3{s.Min.X + float64(u)*constants.LightGrid, s.Min.Y + float64(v)*constants.LightGrid, s.TopZ}
	if floor {
		r.Z = s.BottomZ
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
