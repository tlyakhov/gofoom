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

	Bottom        SectorPlane      `editable:"Bottom"`
	Top           SectorPlane      `editable:"Top"`
	Gravity       concepts.Vector3 `editable:"Gravity"`
	FloorFriction float64          `editable:"Floor Friction"`

	Segments         []*SectorSegment
	Bodies           map[ecs.Entity]*Body
	InternalSegments map[ecs.Entity]*InternalSegment
	// TODO: Should be automatic:
	Inner ecs.EntityTable `editable:"Inner Sectors"`
	Outer ecs.EntityTable

	EnterScripts []*Script `editable:"Enter Scripts"`
	ExitScripts  []*Script `editable:"Exit Scripts"`
	NoShadows    bool      `editable:"No Shadows"`

	Transform dynamic.DynamicValue[concepts.Matrix2] `editable:"Transform"`

	Concave          bool
	Winding          int8
	Min, Max, Center concepts.Vector3

	// Lightmap data
	Lightmap      *xsync.MapOf[uint64, *LightmapCell]
	LightmapBias  [3]int64 // Quantized Min
	LastSeenFrame atomic.Int64
}

func (s *Sector) String() string {
	return "Sector"
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

func (s *Sector) OuterAt(p *concepts.Vector2) (result *Sector) {
	for _, e := range s.Outer {
		if e == 0 {
			continue
		}
		if outer := GetSector(e); outer != nil {
			// This could ensure we pick at least one if there are any
			//			result = outer
			if outer.IsPointInside2D(p) {
				return outer
			}
		}
	}
	return
}

func (s *Sector) ZAt(p *concepts.Vector2) (fz, cz float64) {
	return s.Bottom.ZAt(p), s.Top.ZAt(p)
}

func (s *Sector) removeAdjacentReferences() {
	if s.Attachments == 0 || !s.IsAttached() {
		return
	}

	arena := ecs.ArenaFor[Sector](SectorCID)
	for i := range arena.Cap() {
		sector := arena.Value(i)
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
	if s.IsAttached() {
		s.Top.Z.Detach(ecs.Simulation)
		s.Bottom.Z.Detach(ecs.Simulation)
		s.Transform.Detach(ecs.Simulation)
		s.removeAdjacentReferences()
	}
	for _, b := range s.Bodies {
		b.SectorEntity = 0
	}
	s.Bodies = make(map[ecs.Entity]*Body)
	s.InternalSegments = make(map[ecs.Entity]*InternalSegment)
}

func (s *Sector) OnAttach() {
	s.Attached.OnAttach()
	s.Top.Z.Attach(ecs.Simulation)
	s.Bottom.Z.Attach(ecs.Simulation)
	s.Transform.Attach(ecs.Simulation)
	s.Transform.OnRender = func(blend float64) {
		// TODO: to be able to do this, we need P to be a DynamicValue
	}
	// When we attach a component, its address may change. Ensure segments don't
	// wind up referencing an unattached sector.
	for _, seg := range s.Segments {
		seg.Sector = s
	}
}

func (s *Sector) AddSegment(x float64, y float64) *SectorSegment {
	segment := new(SectorSegment)
	segment.Construct(nil)
	segment.Sector = s
	segment.P = concepts.Vector2{x, y}
	s.Segments = append(s.Segments, segment)
	return segment
}

func (s *Sector) Construct(data map[string]any) {
	s.Attached.Construct(data)

	s.Lightmap = xsync.NewMapOf[uint64, *LightmapCell](xsync.WithPresize(1024))
	s.Segments = make([]*SectorSegment, 0)
	s.Bodies = make(map[ecs.Entity]*Body)
	s.InternalSegments = make(map[ecs.Entity]*InternalSegment)
	s.Gravity[0] = 0
	s.Gravity[1] = 0
	s.Gravity[2] = -constants.Gravity
	s.FloorFriction = 0.85
	s.Bottom.Sector = s
	s.Bottom.Construct(nil)
	s.Bottom.Normal[2] = 1
	s.Bottom.Normal[1] = 0
	s.Bottom.Normal[0] = 0
	s.Top.Sector = s
	s.Top.Construct(nil)
	s.Top.Normal[2] = -1
	s.Top.Normal[1] = 0
	s.Top.Normal[0] = 0
	s.Top.Z.SetAll(64.0)
	s.Transform.Construct(nil)

	if data == nil {
		return
	}

	if v, ok := data["Bottom"]; ok {
		s.Bottom.Construct(v.(map[string]any))
	}
	if v, ok := data["Top"]; ok {
		s.Top.Construct(v.(map[string]any))
	}

	if v, ok := data["Segments"]; ok {
		jsonSegments := v.([]any)
		s.Segments = make([]*SectorSegment, len(jsonSegments))
		for i, jsonSegment := range jsonSegments {
			segment := new(SectorSegment)
			segment.Sector = s
			segment.Construct(jsonSegment.(map[string]any))
			s.Segments[i] = segment
		}
	}

	if v, ok := data["NoShadows"]; ok {
		s.NoShadows = cast.ToBool(v)
	}

	if v, ok := data["Gravity"]; ok {
		s.Gravity.Deserialize(v.(string))
	}
	if v, ok := data["FloorFriction"]; ok {
		s.FloorFriction = cast.ToFloat64(v)
	}

	if v, ok := data["EnterScripts"]; ok {
		s.EnterScripts = ecs.ConstructSlice[*Script](v, nil)
	}
	if v, ok := data["ExitScripts"]; ok {
		s.ExitScripts = ecs.ConstructSlice[*Script](v, nil)
	}
	if v, ok := data["Inner"]; ok {
		s.Inner = ecs.ParseEntityTable(v, false)
	}
	if v, ok := data["Transform"]; ok {
		s.Transform.Construct(v.(map[string]any))
	}

	s.Recalculate()
}

func (s *Sector) Serialize() map[string]any {
	result := s.Attached.Serialize()
	result["Top"] = s.Top.Serialize()
	result["Bottom"] = s.Bottom.Serialize()
	result["FloorFriction"] = s.FloorFriction
	if s.NoShadows {
		result["NoShadows"] = s.NoShadows
	}

	if s.Gravity[0] != 0 || s.Gravity[1] != 0 || s.Gravity[2] != -constants.Gravity {
		result["Gravity"] = s.Gravity.Serialize()
	}
	if len(s.EnterScripts) > 0 {
		result["EnterScripts"] = ecs.SerializeSlice(s.EnterScripts)
	}
	if len(s.ExitScripts) > 0 {
		result["ExitScripts"] = ecs.SerializeSlice(s.ExitScripts)
	}

	if !s.Inner.Empty() {
		result["Inner"] = s.Inner.Serialize()
	}

	result["Transform"] = s.Transform.Serialize()

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
		bz, tz := s.ZAt(&segment.P)
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

	s.Top.Recalculate()
	s.Bottom.Recalculate()

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

	for _, e := range s.Inner {
		if e == 0 {
			continue
		}
		if inner := GetSector(e); inner != nil {
			inner.Outer.Set(s.Entity)
		}
	}

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

func (s *Sector) InnermostContaining(p *concepts.Vector2) (result *Sector) {
	if s == nil {
		return nil
	} else if s.IsPointInside2D(p) {
		result = s
	} else {
		return nil
	}

	for _, e := range s.Inner {
		if e == 0 {
			continue
		}
		inner := GetSector(e)
		if inner = inner.InnermostContaining(p); inner != nil {
			result = inner
		}
	}
	return
}
