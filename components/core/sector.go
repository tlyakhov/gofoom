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

	Layer        int `editable:"Layer"` // TODO: Add more editor support
	HigherLayers ecs.EntityTable
	LowerLayers  ecs.EntityTable

	EnterScripts []*Script `editable:"Enter Scripts"`
	ExitScripts  []*Script `editable:"Exit Scripts"`
	NoShadows    bool      `editable:"No Shadows"`

	Transform       dynamic.DynamicValue[concepts.Matrix2] `editable:"Transform"`
	TransformOrigin concepts.Vector2                       `editable:"Transform Origin"`

	Concave  bool
	Winding  int8
	Min, Max concepts.Vector3
	Center   dynamic.DynamicValue[concepts.Vector3]

	// Lightmap data
	Lightmap      *xsync.MapOf[uint64, *LightmapCell]
	LightmapBias  [3]int64 // Quantized Min
	LastSeenFrame atomic.Int64
}

func (s *Sector) String() string {
	return "Sector"
}

func (s *Sector) IsPointInside2D(p *concepts.Vector2) bool {
	if len(s.Segments) == 0 {
		return false
	}

	inside := false
	flag1 := (p[1] >= s.Segments[0].P.Render[1])

	for _, segment := range s.Segments {
		flag2 := (p[1] >= segment.Next.P.Render[1])
		if flag1 != flag2 {
			if ((segment.Next.P.Render[1]-p[1])*(segment.P.Render[0]-segment.Next.P.Render[0]) >=
				(segment.Next.P.Render[0]-p[0])*(segment.P.Render[1]-segment.Next.P.Render[1])) == flag2 {
				inside = !inside
			}
		}
		flag1 = flag2
	}
	return inside
}

func (s *Sector) OverlapAt(p *concepts.Vector2, lower bool) (result *Sector) {
	overlaps := s.HigherLayers
	if lower {
		overlaps = s.LowerLayers
	}
	for _, e := range overlaps {
		if e == 0 {
			continue
		}
		if overlap := GetSector(e); overlap != nil {
			if overlap.IsPointInside2D(p) {
				return overlap
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
		s.Center.Detach(ecs.Simulation)
		s.removeAdjacentReferences()
		for _, seg := range s.Segments {
			seg.P.Detach(ecs.Simulation)
		}
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
	s.Center.Attach(ecs.Simulation)

	// When we attach a component, its address may change. Ensure segments don't
	// wind up referencing an unattached sector.
	for _, seg := range s.Segments {
		seg.Sector = s
		seg.P.Attach(ecs.Simulation)
	}
}

func (s *Sector) AddSegment(x float64, y float64) *SectorSegment {
	segment := new(SectorSegment)
	segment.Construct(nil)
	segment.Sector = s
	segment.P.SetAll(concepts.Vector2{x, y})
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
	s.Layer = 0

	if data == nil {
		return
	}

	if v, ok := data["Layer"]; ok {
		s.Layer = cast.ToInt(v)
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
			segment.P.Attach(ecs.Simulation)
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
	if v, ok := data["Transform"]; ok {
		s.Transform.Construct(v)
	}
	if v, ok := data["TransformOrigin"]; ok {
		s.TransformOrigin.Deserialize(v.(string))
	}

	s.Recalculate()
}

func (s *Sector) Serialize() map[string]any {
	result := s.Attached.Serialize()
	if s.Layer != 0 {
		result["Layer"] = s.Layer
	}
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

	result["Transform"] = s.Transform.Serialize()
	result["TransformOrigin"] = s.TransformOrigin.Serialize()

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

func (s *Sector) RecalculateNonTopological() {
	s.Center.SetAll(concepts.Vector3{0, 0, 0})
	s.Min[0] = math.Inf(1)
	s.Min[1] = math.Inf(1)
	s.Min[2] = math.Inf(1)
	s.Max[0] = math.Inf(-1)
	s.Max[1] = math.Inf(-1)
	s.Max[2] = math.Inf(-1)

	for _, seg := range s.Segments {
		// TODO: Maybe optimize this by not recalculating spawn values unless
		// points have changed? maybe too complicated
		s.Center.Now[0] += seg.P.Render[0]
		s.Center.Now[1] += seg.P.Render[1]
		s.Center.Spawn[0] += seg.P.Spawn[0]
		s.Center.Spawn[1] += seg.P.Spawn[1]
		if seg.P.Render[0] < s.Min[0] {
			s.Min[0] = seg.P.Render[0]
		}
		if seg.P.Render[1] < s.Min[1] {
			s.Min[1] = seg.P.Render[1]
		}
		if seg.P.Render[0] > s.Max[0] {
			s.Max[0] = seg.P.Render[0]
		}
		if seg.P.Render[1] > s.Max[1] {
			s.Max[1] = seg.P.Render[1]
		}
		bz, tz := s.ZAt(&seg.P.Render)
		s.Center.Now[2] += (bz + tz) * 0.5
		sbz, stz := s.ZAt(&seg.P.Spawn)
		s.Center.Spawn[2] += (sbz + stz) * 0.5
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
		seg.Recalculate()
	}

	s.Center.Now.MulSelf(1.0 / float64(len(s.Segments)))
	s.Center.Spawn.MulSelf(1.0 / float64(len(s.Segments)))
	s.LightmapBias[0] = math.MaxInt64

	s.Top.Recalculate()
	s.Bottom.Recalculate()

	s.HigherLayers = ecs.EntityTable{}
	s.LowerLayers = ecs.EntityTable{}
	arena := ecs.ArenaFor[Sector](SectorCID)
	// TODO: Optimize this to not have to iterate all of the sectors. Easiest
	// would probably be to use the quad tree.
	for i := range arena.Cap() {
		test := arena.Value(i)
		// Ignore overlaps that have the same layer
		if test == nil || s.Layer == test.Layer {
			continue
		}
		if s.AABBIntersect(&test.Min, &test.Max, true) {
			if test.Layer > s.Layer {
				s.HigherLayers.Set(test.Entity)
				test.LowerLayers.Set(s.Entity)
			} else {
				s.LowerLayers.Set(test.Entity)
				test.HigherLayers.Set(s.Entity)
			}
		}
	}
}

func (s *Sector) Recalculate() {
	// TODO: Make this method more efficient
	sum := 0.0
	for i, segment := range s.Segments {
		// Can't use prev/next pointers because they haven't been initialized yet.
		next := s.Segments[(i+1)%len(s.Segments)]
		sum += (next.P.Render[0] - segment.P.Render[0]) * (segment.P.Render[1] + next.P.Render[1])
	}

	if sum < 0 {
		s.Winding = 1
	} else {
		s.Winding = -1
	}

	for i, seg := range s.Segments {
		next := s.Segments[(i+1)%len(s.Segments)]
		seg.Index = i
		seg.Next = next
		next.Prev = seg
		seg.Sector = s
	}

	if len(s.Segments) > 1 {
		// Figure out if this sector is concave.
		// The algorithm tests if any angles are > 180 degrees
		prev := 0.0
		for _, s1 := range s.Segments {
			s2 := s1.Next
			s3 := s2.Next
			d1 := s2.P.Render.Sub(&s1.P.Render)
			d2 := s3.P.Render.Sub(&s2.P.Render)
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

	s.RecalculateNonTopological()

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

// From https://web.archive.org/web/20100405070507/http://valis.cs.uiuc.edu/~sariel/research/CG/compgeom/msg00831.html
func (s *Sector) Area() float64 {
	sum := 0.0
	for _, seg := range s.Segments {
		sum += seg.P.Render[0] * seg.Next.P.Render[1]
		sum -= seg.P.Render[1] * seg.Next.P.Render[0]
	}
	return math.Abs(sum) * 0.5
}

func (s *Sector) DbgPrintSegments() {
	log.Printf("Segments:")
	for i, seg := range s.Segments {
		log.Printf("%v: %v", i, seg.P.Render.StringHuman())
	}
}

func (s *Sector) TopmostSector(p *concepts.Vector2) (result *Sector) {
	if s == nil {
		return nil
	} else if s.IsPointInside2D(p) {
		result = s
	} else {
		return nil
	}

	for _, e := range s.HigherLayers {
		if e == 0 {
			continue
		}
		overlap := GetSector(e)
		if overlap = overlap.TopmostSector(p); overlap != nil {
			result = overlap
		}
	}
	return
}

func (s *Sector) Contains2D(s2 *Sector) bool {
	if !s.AABBIntersect2D(s2.Min.To2D(), s2.Max.To2D(), true) {
		return false
	}
	r := concepts.Vector2{}
	for _, seg2 := range s2.Segments {
		for _, seg := range s.Segments {
			if seg.Intersect2D(&seg2.P.Render, &seg2.Next.P.Render, &r) >= 0 {
				return false
			}
		}
		if !s.IsPointInside2D(&seg2.P.Render) {
			return false
		}
	}

	return true
}
