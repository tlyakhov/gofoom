package mapping

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
	LightmapWidth, LightmapHeight uint
	FloorLightmap, CeilLightmap   []float64
	PVSEntity                     map[string]PhysicalSector
	// RoomImpulse
	// PVS
	// PVSLights []
}

type AbstractSector interface {
	concepts.ISerializable
	GetPhysical() *PhysicalSector
	IsPointInside2D(p *concepts.Vector2) bool
}

func init() {
	registry.Instance().Register(PhysicalSector{})
}

func (s *PhysicalSector) GetPhysical() *PhysicalSector {
	return s
}

func (ms *PhysicalSector) Recalculate() {
	ms.Center = &concepts.Vector3{0, 0, (ms.TopZ + ms.BottomZ) / 2}
	ms.Min = &concepts.Vector3{math.Inf(1), math.Inf(1), ms.BottomZ}
	ms.Max = &concepts.Vector3{math.Inf(-1), math.Inf(-1), ms.TopZ}

	w := ms.Winding()

	for i, segment := range ms.Segments {
		next := ms.Segments[(i+1)%len(ms.Segments)]
		ms.Center.X += segment.A.X
		ms.Center.Y += segment.A.Y
		if segment.A.X < ms.Min.X {
			ms.Min.X = segment.A.X
		}
		if segment.A.Y < ms.Min.Y {
			ms.Min.Y = segment.A.Y
		}
		if segment.A.X > ms.Max.X {
			ms.Max.X = segment.A.X
		}
		if segment.A.Y > ms.Max.Y {
			ms.Max.Y = segment.A.Y
		}
		segment.Sector = ms
		segment.B = next.A
		segment.Recalculate()

		if !w {
			segment.Normal = segment.Normal.Mul(-1)
		}
	}

	ms.Center = ms.Center.Mul(1.0 / float64(len(ms.Segments)))

	for _, e := range ms.Entities {
		e.GetPhysical().Map = ms.Map
		e.GetPhysical().Sector = ms
		if c, ok := e.(Collideable); ok {
			c.Collide()
		}
	}

	ms.LightmapWidth = uint((ms.Max.X-ms.Min.X)/constants.LightGrid) + 6
	ms.LightmapHeight = uint((ms.Max.Y-ms.Min.Y)/constants.LightGrid) + 6
	ms.FloorLightmap = make([]float64, ms.LightmapWidth*ms.LightmapHeight*3)
	ms.CeilLightmap = make([]float64, ms.LightmapWidth*ms.LightmapHeight*3)
	ms.ClearLightmaps()
}

func (ms *PhysicalSector) ClearLightmaps() {
	for i := range ms.FloorLightmap {
		ms.FloorLightmap[i] = -1
		ms.CeilLightmap[i] = -1
	}

	for _, segment := range ms.Segments {
		segment.ClearLightmap()
	}

	// ms.UpdatePVS()
	// ms.UpdateEntityPVS()
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

func (s *PhysicalSector) SetParent(parent interface{}) {
	if m, ok := parent.(*Map); ok {
		s.Map = m
	} else {
		panic("Tried mapping.PhysicalSector.SetParent with a parameter that wasn't a *mapping.Map")
	}
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
	s.Recalculate()
}
