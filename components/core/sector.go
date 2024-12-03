// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package core

import (
	"log"
	"math"
	"sync/atomic"

	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/dynamic"
	"tlyakhov/gofoom/ecs"

	"github.com/puzpuzpuz/xsync/v3"
	"github.com/spf13/cast"
)

type Sector struct {
	ecs.Attached `editable:"^"`

	Bottom           SectorPlane      `editable:"Bottom"`
	Top              SectorPlane      `editable:"Top"`
	Gravity          concepts.Vector3 `editable:"Gravity"`
	FloorFriction    float64          `editable:"Floor Friction"`
	Segments         []*SectorSegment
	Bodies           map[ecs.Entity]*Body
	Colliders        map[ecs.Entity]*Mobile
	InternalSegments map[ecs.Entity]*InternalSegment
	EnterScripts     []*Script `editable:"Enter Scripts"`
	ExitScripts      []*Script `editable:"Exit Scripts"`

	Concave          bool
	Winding          int8
	Min, Max, Center concepts.Vector3

	// Potentially visible set
	LastPVSRefresh uint64
	PVS            map[ecs.Entity]*Sector
	// Potentially visible lights
	PVL []*Body

	// Lightmap data
	Lightmap      *xsync.MapOf[uint64, *LightmapCell]
	LightmapBias  [3]int64 // Quantized Min
	LastSeenFrame atomic.Int64
}

var SectorCID ecs.ComponentID

func init() {
	SectorCID = ecs.RegisterComponent(&ecs.Column[Sector, *Sector]{Getter: GetSector})
}

func GetSector(db *ecs.ECS, e ecs.Entity) *Sector {
	if asserted, ok := db.Component(e, SectorCID).(*Sector); ok {
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

func (s *Sector) ZAt(stage dynamic.DynamicStage, p *concepts.Vector2) (fz, cz float64) {
	return s.Bottom.ZAt(stage, p), s.Top.ZAt(stage, p)
}

func (s *Sector) removeAdjacentReferences() {
	if s.Attachments == 0 || s.ECS == nil {
		return
	}

	col := ecs.ColumnFor[Sector](s.ECS, SectorCID)
	for i := range col.Cap() {
		sector := col.Value(i)
		if sector == nil {
			continue
		}
		for _, seg := range sector.Segments {
			if s.Entities.Contains(seg.AdjacentSector) {
				seg.AdjacentSector = 0
				seg.AdjacentSegment = nil
				seg.AdjacentSegmentIndex = 0
			}
		}
	}
}
func (s *Sector) OnDelete() {
	defer s.Attached.OnDelete()
	if s.ECS != nil {
		s.Top.Z.Detach(s.ECS.Simulation)
		s.Bottom.Z.Detach(s.ECS.Simulation)
		s.removeAdjacentReferences()
	}
	for _, b := range s.Bodies {
		b.SectorEntity = 0
	}
	s.Bodies = make(map[ecs.Entity]*Body)
	s.Colliders = make(map[ecs.Entity]*Mobile)
	s.InternalSegments = make(map[ecs.Entity]*InternalSegment)
}

func (s *Sector) OnAttach(db *ecs.ECS) {
	s.Attached.OnAttach(db)
	s.Top.Z.Attach(db.Simulation)
	s.Bottom.Z.Attach(db.Simulation)
}

func (s *Sector) AddSegment(x float64, y float64) *SectorSegment {
	segment := new(SectorSegment)
	segment.Construct(s.ECS, nil)
	segment.Sector = s
	segment.P = concepts.Vector2{x, y}
	s.Segments = append(s.Segments, segment)
	return segment
}

func (s *Sector) Construct(data map[string]any) {
	s.Attached.Construct(data)

	s.PVS = make(map[ecs.Entity]*Sector)
	s.PVL = make([]*Body, 0)
	s.Lightmap = xsync.NewMapOf[uint64, *LightmapCell](xsync.WithPresize(1024))
	s.Segments = make([]*SectorSegment, 0)
	s.Bodies = make(map[ecs.Entity]*Body)
	s.Colliders = make(map[ecs.Entity]*Mobile)
	s.InternalSegments = make(map[ecs.Entity]*InternalSegment)
	s.Gravity[0] = 0
	s.Gravity[1] = 0
	s.Gravity[2] = -constants.Gravity
	s.FloorFriction = 0.85
	s.Bottom.Construct(s, nil)
	s.Bottom.Normal[2] = 1
	s.Bottom.Normal[1] = 0
	s.Bottom.Normal[0] = 0
	s.Top.Construct(s, nil)
	s.Top.Normal[2] = -1
	s.Top.Normal[1] = 0
	s.Top.Normal[0] = 0
	s.Top.Z.SetAll(64.0)

	if data == nil {
		return
	}

	// TODO: Remove the following after all the world files are migrated
	if _, ok := data["Bottom"]; !ok {
		data["Bottom"] = make(map[string]any)
	}
	if _, ok := data["Top"]; !ok {
		data["Top"] = make(map[string]any)
	}
	if v, ok := data["Bottom.Z"]; ok {
		data["Bottom"].(map[string]any)["Z"] = v
	}
	if v, ok := data["Top.Z"]; ok {
		data["Top"].(map[string]any)["Z"] = v
	}
	if v, ok := data["Bottom.Normal"]; ok {
		data["Bottom"].(map[string]any)["Normal"] = v
	}
	if v, ok := data["Top.Normal"]; ok {
		data["Top"].(map[string]any)["Normal"] = v
	}
	if v, ok := data["Bottom.Target"]; ok {
		data["Bottom"].(map[string]any)["Target"] = v
	}
	if v, ok := data["Top.Target"]; ok {
		data["Top"].(map[string]any)["Target"] = v
	}
	if v, ok := data["Bottom.Surface"]; ok {
		data["Bottom"].(map[string]any)["Surface"] = v
	}
	if v, ok := data["Top.Surface"]; ok {
		data["Top"].(map[string]any)["Surface"] = v
	}
	if v, ok := data["Bottom.Scripts"]; ok {
		data["Bottom"].(map[string]any)["Scripts"] = v
	}
	if v, ok := data["Top.Scripts"]; ok {
		data["Top"].(map[string]any)["Scripts"] = v
	}
	// END TODO

	if v, ok := data["Bottom"]; ok {
		s.Bottom.Construct(s, v.(map[string]any))
	}
	if v, ok := data["Top"]; ok {
		s.Top.Construct(s, v.(map[string]any))
	}

	if v, ok := data["Segments"]; ok {
		jsonSegments := v.([]any)
		s.Segments = make([]*SectorSegment, len(jsonSegments))
		for i, jsonSegment := range jsonSegments {
			segment := new(SectorSegment)
			segment.Sector = s
			segment.Construct(s.ECS, jsonSegment.(map[string]any))
			s.Segments[i] = segment
		}
	}

	if v, ok := data["Gravity"]; ok {
		s.Gravity.Deserialize(v.(string))
	}
	if v, ok := data["FloorFriction"]; ok {
		s.FloorFriction = cast.ToFloat64(v)
	}

	if v, ok := data["EnterScripts"]; ok {
		s.EnterScripts = ecs.ConstructSlice[*Script](s.ECS, v, nil)
	}
	if v, ok := data["ExitScripts"]; ok {
		s.ExitScripts = ecs.ConstructSlice[*Script](s.ECS, v, nil)
	}

	s.Recalculate()
}

func (s *Sector) Serialize() map[string]any {
	result := s.Attached.Serialize()
	result["Top"] = s.Top.Serialize()
	result["Bottom"] = s.Bottom.Serialize()
	result["FloorFriction"] = s.FloorFriction

	if s.Gravity[0] != 0 || s.Gravity[1] != 0 || s.Gravity[2] != -constants.Gravity {
		result["Gravity"] = s.Gravity.Serialize()
	}
	if len(s.EnterScripts) > 0 {
		result["EnterScripts"] = ecs.SerializeSlice(s.EnterScripts)
	}
	if len(s.ExitScripts) > 0 {
		result["ExitScripts"] = ecs.SerializeSlice(s.ExitScripts)
	}

	segments := []any{}
	for _, seg := range s.Segments {
		segments = append(segments, seg.Serialize())
	}
	result["Segments"] = segments
	return result
}

var contactScriptParams = []ScriptParam{
	{Name: "body", TypeName: "*core.Body"},
	{Name: "sector", TypeName: "*core.Sector"},
}

func (s *Sector) Recalculate() {
	// TODO: Make this method more efficient
	concepts.V3(&s.Center, 0, 0, 0)
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

	for i, segment := range s.Segments {
		next := s.Segments[(i+1)%len(s.Segments)]
		segment.Index = i
		segment.Next = next
		next.Prev = segment
		//prev = segment
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
		bz, tz := s.ZAt(dynamic.DynamicSpawn, &segment.P)
		s.Center[2] += (bz + tz) * 0.5
		if bz < s.Min[2] {
			s.Min[2] = bz
		}
		if tz < s.Min[2] {
			s.Min[2] = tz
		}
		if bz > s.Max[2] {
			s.Max[2] = bz
		}
		if tz > s.Max[2] {
			s.Max[2] = tz
		}
		segment.Sector = s
		segment.Recalculate()
	}

	if len(s.Segments) > 1 {
		// Figure out if this sector is concave.
		// The algorithm tests if any angles are > 180 degrees
		prev := 0.0
		for _, s1 := range s.Segments {
			s2 := s1.Next
			s3 := s2.Next
			d1 := s2.P.Sub(&s1.P)
			d2 := s3.P.Sub(&s2.P)
			c := d1.Cross(d2)
			if c != 0 {
				if c*prev < 0 {
					s.Concave = true
					break
				}
				prev = c
			}
		}
	}

	s.Center.MulSelf(1.0 / float64(len(s.Segments)))
	s.LightmapBias[0] = math.MaxInt64
	for _, script := range s.EnterScripts {
		script.Params = contactScriptParams
		script.Compile()
	}
	for _, script := range s.ExitScripts {
		script.Params = contactScriptParams
		script.Compile()
	}
	for _, script := range s.Top.Scripts {
		script.Params = contactScriptParams
		script.Compile()
	}
	for _, script := range s.Bottom.Scripts {
		script.Params = contactScriptParams
		script.Compile()
	}
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
