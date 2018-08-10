package core

import (
	"math"

	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/constants"
	"github.com/tlyakhov/gofoom/registry"
)

type PhysicalSector struct {
	concepts.Base

	Map                     *Map
	Segments                []*Segment
	Entities                map[string]AbstractEntity
	BottomZ, TopZ           float64
	FloorScale, CeilScale   float64
	FloorTarget, CeilTarget AbstractSector
	FloorMaterial           concepts.ISerializable
	CeilMaterial            concepts.ISerializable

	Min, Max, Center              *concepts.Vector3
	LightmapWidth, LightmapHeight uint32
	FloorLightmap, CeilLightmap   []float64
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

func (ms *PhysicalSector) IsPointInside2D(p *concepts.Vector2) bool {
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

func (s *PhysicalSector) Recalculate() {
	s.Center = &concepts.Vector3{0, 0, (s.TopZ + s.BottomZ) / 2}
	s.Min = &concepts.Vector3{math.Inf(1), math.Inf(1), s.BottomZ}
	s.Max = &concepts.Vector3{math.Inf(-1), math.Inf(-1), s.TopZ}

	w := s.Winding()

	for i, segment := range s.Segments {
		next := s.Segments[(i+1)%len(s.Segments)]
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

	s.LightmapWidth = uint32((s.Max.X-s.Min.X)/constants.LightGrid) + 6
	s.LightmapHeight = uint32((s.Max.Y-s.Min.Y)/constants.LightGrid) + 6
	s.FloorLightmap = make([]float64, s.LightmapWidth*s.LightmapHeight*3)
	s.CeilLightmap = make([]float64, s.LightmapWidth*s.LightmapHeight*3)
	s.ClearLightmaps()
}

func (s *PhysicalSector) ClearLightmaps() {
	for i := range s.FloorLightmap {
		s.FloorLightmap[i] = -1
		s.CeilLightmap[i] = -1
	}

	for _, segment := range s.Segments {
		segment.ClearLightmap()
	}
}

func (s *PhysicalSector) LightmapAddress(p *concepts.Vector2) uint32 {
	dx := int(p.X-s.Min.X)/constants.LightGrid + 3
	dy := int(p.Y-s.Min.Y)/constants.LightGrid + 3
	if dx < 0 {
		dx = 0
	}
	if dy < 0 {
		dy = 0
	}
	return uint32((uint32(dy)*s.LightmapWidth + uint32(dx)) * 3)
}

func (s *PhysicalSector) LightmapWorld(p *concepts.Vector3) *concepts.Vector3 {
	lw := p.Sub(s.Min).Mul(1.0 / constants.LightGrid)
	lw.X = math.Floor(lw.X) * constants.LightGrid
	lw.Y = math.Floor(lw.Y) * constants.LightGrid
	lw.Z = p.Z
	return lw
}

func (s *PhysicalSector) LightmapAddressToWorld(mapIndex uint32, floor bool) *concepts.Vector3 {
	u := int((mapIndex/3)%s.LightmapWidth) - 3
	v := int((mapIndex/3)/s.LightmapWidth) - 3
	r := concepts.Vector3{s.Min.X + float64(u)*constants.LightGrid, s.Min.Y + float64(v)*constants.LightGrid, s.TopZ}
	if floor {
		r.Z = s.BottomZ
	}
	return &r
}

func (s *PhysicalSector) SetParent(parent interface{}) {
	if m, ok := parent.(*Map); ok {
		s.Map = m
	} else {
		panic("Tried mapping.PhysicalSector.SetParent with a parameter that wasn't a *mapping.Map")
	}
}
