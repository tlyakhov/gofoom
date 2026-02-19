// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package render

import (
	"log"
	"math"
	"sync/atomic"

	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/ecs"
)

const LogDebug = false

// For glitchTester5.yml (vertical shadow line)
// const LogDebugLightHash = 0x41005900097a40
// For inner-test.yml (diagonal shadow line)
// const LogDebugLightHash = 0x1c000f00064400
// For hall.yml (black ceil slope in hallway)
// const LogDebugLightHash = 0x1900320013600c
// For glitchTester6.yml (weird flickering shadows on edges)
// const LogDebugLightHash = 0x4000270000000f
const LogDebugLightHash = 0x4100280000000f

const LogDebugLightEntity = 4

const PackedLightBits = 10
const PackedLightMax = (1 << PackedLightBits) - 1
const PackedLightRange = 12

// All the data and state required to retrieve/calculate a lightmap voxel
type LightSampler struct {
	MaterialSampler
	Hash             uint64
	Output           concepts.Vector3
	Filter           concepts.Vector4
	Hit              concepts.Vector3
	IntersectionTest concepts.Vector3
	Q                concepts.Vector3
	LightWorld       concepts.Vector3
	InputBody        ecs.Entity
	Visited          []*core.Sector

	Sector  *core.Sector
	Segment *core.Segment
	// This will be different from .Sector for inner segments.
	SegmentSector *core.Sector
	Normal        concepts.Vector3

	Intersection core.RayIntersection

	prevDistSq, hitDistSq, maxDistSq float64

	maxDist float64

	xorSeed uint64
}

func (ls *LightSampler) packLight() uint32 {
	return uint32(ls.Output[0]*PackedLightMax/PackedLightRange)<<(PackedLightBits+PackedLightBits) |
		uint32(ls.Output[1]*PackedLightMax/PackedLightRange)<<PackedLightBits |
		uint32(ls.Output[2]*PackedLightMax/PackedLightRange)
}

func (ls *LightSampler) unpackLight(c uint32) {
	ls.Output[0] = float64((c>>(PackedLightBits+PackedLightBits))&PackedLightMax) * PackedLightRange / PackedLightMax
	ls.Output[1] = float64((c>>PackedLightBits)&PackedLightMax) * PackedLightRange / PackedLightMax
	ls.Output[2] = float64(c&PackedLightMax) * PackedLightRange / PackedLightMax
}

func (ls *LightSampler) Debug() *concepts.Vector3 {
	ls.LightmapHashToWorld(ls.Sector, &ls.Q, ls.Hash)
	dbg := ls.Q.Mul(1.0 / 64.0)
	ls.Output[0] = dbg[0] - math.Floor(dbg[0])
	ls.Output[1] = dbg[1] - math.Floor(dbg[1])
	ls.Output[2] = dbg[2] - math.Floor(dbg[2])
	return &ls.Output
}

func (ls *LightSampler) Get() *concepts.Vector3 {
	if lmResult, exists := ls.Sector.Lightmap.Load(ls.Hash); exists {
		r := concepts.RngXorShift64(ls.xorSeed ^ ls.Hash ^ uint64(ls.ScreenY))
		if lmResult.Timestamp+constants.MaxLightmapAge >= uint32(ecs.Simulation.Frame) ||
			!concepts.RngDecide(r, constants.LightmapRefreshDither) {
			ls.unpackLight(lmResult.Light)
			/*	if LogDebug && LogDebugLightHash == ls.Hash {
				log.Printf("Lightmap value exists: %v\n", ls.Output.StringHuman(2))
			}*/
			return &ls.Output
		}
	}
	ls.LightmapHashToWorld(ls.Sector, &ls.Q, ls.Hash)
	// Ensure our quantized world location is within Z bounds to avoid
	// weird shadowing.
	fz, cz := ls.Sector.ZAt(ls.Q.To2D())
	if ls.Q[2] < fz {
		ls.Q[2] = fz
	}
	if ls.Q[2] > cz {
		ls.Q[2] = cz
	}
	ls.ScaleW = 64
	ls.ScaleH = 64
	ls.Calculate(&ls.Q)
	ls.Sector.Lightmap.Store(ls.Hash, &core.LightmapCell{
		Light:     ls.packLight(),
		Timestamp: uint32(ecs.Simulation.Frame),
	})
	return &ls.Output

}

// lightVisible determines whether a given light is visible from a world location.
func (ls *LightSampler) lightVisible(p *concepts.Vector3, body *core.Body) bool {
	if ls.Sector.NoShadows {
		return true
	}

	p2d := p.To2D()

	// Check higher level sectors - this is valuable for edge cases where our voxel
	// is near a boundary with an inner sector
	for _, e := range ls.Sector.HigherLayers {
		if e == 0 {
			continue
		}
		overlap := core.GetSector(e)
		// If our light sample is on the segment itself, don't shadow it.
		if ls.SegmentSector == overlap && ls.Normal[2] == 0 {
			continue
		}

		// Check if the lightmap sample is inside the overlapped sector.
		if !overlap.IsPointInside2D(p2d) {
			continue
		}
		// If we're here, our lightmap sample is inside a higher layer sector,
		// If it's under the floor or above the ceiling, it's shadowed for sure.
		floorZ, ceilZ := overlap.ZAt(p2d)
		if p[2]-ceilZ > ls.LightGrid || floorZ-p[2] > ls.LightGrid {
			return false
		}
		// TODO: This should be recursive, probably?
		if ls.lightVisibleFromSector(p, body, overlap) {
			return true
		}
	}

	// Check the starting sector
	if ls.lightVisibleFromSector(p, body, ls.Sector) {
		return true
	}

	// Check exterior sectors in case our lighting sample is just outside the
	// test sector. First, check adjacencies:
	for _, seg := range ls.Sector.Segments {
		if seg.AdjacentSector == 0 || seg.AdjacentSegment == nil || seg.PortalHasMaterial {
			continue
		}
		distSq := seg.AdjacentSegment.DistanceToPointSq(p2d)
		if distSq >= ls.LightGrid*ls.LightGrid*2 {
			continue
		}

		floorZ, ceilZ := seg.AdjacentSegment.Sector.ZAt(p2d)
		if p[2]-ceilZ > ls.LightGrid || floorZ-p[2] > ls.LightGrid {
			continue
		}
		if ls.lightVisibleFromSector(p, body, seg.AdjacentSegment.Sector) {
			return true
		}
	}
	// Check lower level sectors, maybe we escaped outside of an inner sector:
	for _, e := range ls.Sector.LowerLayers {
		if e == 0 {
			continue
		}
		overlap := core.GetSector(e)
		if !overlap.IsPointInside2D(p2d) {
			continue
		}
		floorZ, ceilZ := overlap.ZAt(p2d)
		if p[2]-ceilZ > ls.LightGrid || floorZ-p[2] > ls.LightGrid {
			return false
		}
		if ls.lightVisibleFromSector(p, body, overlap) {
			return true
		}
	}
	return false
}

// lightVisibleFromSector determines whether a given light is visible from a world location.
func (ls *LightSampler) lightVisibleFromSector(p *concepts.Vector3, lightBody *core.Body, sector *core.Sector) bool {
	if LogDebug && LogDebugLightHash == ls.Hash && LogDebugLightEntity == lightBody.Entity {
		log.Printf("LightSampler.lightVisibleFromSector: START")
		log.Printf("world=%v, light=%v\n", p.String(), lightBody.Pos.Render.String())
	}
	var next, overlap *core.Sector
	var hitSegment *core.SectorSegment

	if ls.maxDist < 0 {
		ls.maxDist = math.Sqrt(ls.maxDistSq)
	}

	sizeSq := lightBody.Size.Render[0] * 0.5
	sizeSq *= sizeSq

	// Initialize ray struct
	ls.Intersection.Start = *p
	ls.Intersection.End = lightBody.Pos.Render
	ls.Intersection.Delta = ls.LightWorld
	ls.Intersection.Limit = ls.maxDist
	ls.Intersection.IgnoreSegment = ls.Segment

	// The outer loop traverses portals starting from the sector our target point is in,
	// and finishes in the sector our light is in (unless occluded)
	depth := 0 // We keep track of portaling depth to avoid infinite traversal in weird cases.
	ls.Visited = ls.Visited[:0]
	ls.prevDistSq = -1.0
	for sector != nil {
		// Since our sectors can be concave or have inner sectors (holes), we
		// can't just go through the first portal we find, we have to go through
		// the NEAREST one. Use hitDistSq to keep track...
		ls.hitDistSq = ls.maxDistSq
		ls.Intersection.MinDistSq = ls.prevDistSq
		ls.Intersection.MaxDistSq = ls.maxDistSq
		ls.Intersection.CheckEntry = false

		if LogDebug && LogDebugLightHash == ls.Hash && LogDebugLightEntity == lightBody.Entity {
			log.Printf("  Checking sector %v, max dist %v", sector.Entity, math.Sqrt(ls.maxDistSq))
		}
		//Intersect this sector
		sector.IntersectRay(&ls.Intersection)

		if ls.Intersection.HitSegment != nil {
			ls.hitDistSq = ls.Intersection.HitDistSq
			ls.Hit = ls.Intersection.HitPoint
			if math.Abs(ls.Intersection.HitDistSq-ls.maxDistSq) < sizeSq {
				hitSegment = nil
				next = nil
			} else {
				if ls.Intersection.NextSector == nil {
					return false // Occluded by wall
				}
				hitSegment = ls.Intersection.HitSegment
				next = ls.Intersection.NextSector
			}
		} else {
			hitSegment = nil
			next = nil
		}

		// Check higher layer sectors for intersections
		ls.Intersection.MaxDistSq = ls.hitDistSq
		ls.Intersection.CheckEntry = true
		for _, e := range sector.HigherLayers {
			if e == 0 {
				continue
			}
			if LogDebug && LogDebugLightHash == ls.Hash && LogDebugLightEntity == lightBody.Entity {
				log.Printf("  Visiting higher layer sector %v, max dist %v", e, ls.maxDistSq)
			}
			if overlap = core.GetSector(e); overlap == nil {
				continue
			}

			overlap.IntersectRay(&ls.Intersection)

			if ls.Intersection.HitSegment != nil {
				ls.hitDistSq = ls.Intersection.HitDistSq
				ls.Hit = ls.Intersection.HitPoint
				if math.Abs(ls.Intersection.HitDistSq-ls.maxDistSq) < sizeSq {
					hitSegment = nil
					next = nil
				} else {
					if ls.Intersection.NextSector == nil {
						return false // Occluded by HigherLayer
					}
					hitSegment = ls.Intersection.HitSegment
					next = ls.Intersection.NextSector
				}
			}
		}
		if next == nil {
			// No portal sectors, not occluded
			ls.Visited = append(ls.Visited, sector)
			break
		}
		// If the portal has a transparent material, we need to filter the light
		if hitSegment.PortalHasMaterial {
			i2d := ls.Hit.To2D()
			floorZ, ceilZ := sector.ZAt(i2d)
			ls.Initialize(hitSegment.Surface.Material, hitSegment.Surface.ExtraStages)
			ls.NU = ls.Hit.To2D().Dist(&hitSegment.P.Render) / hitSegment.Length
			ls.NV = (ceilZ - ls.Hit[2]) / (ceilZ - floorZ)
			ls.U = ls.NU
			ls.V = ls.NV
			ls.SampleMaterial(hitSegment.Surface.ExtraStages)
			if lit := materials.GetLit(hitSegment.Surface.Material); lit != nil {
				lit.Apply(&ls.MaterialSampler.Output, nil)
			}
			if ls.MaterialSampler.Output[3] >= 0.99 {
				return false
			}
			//		log.Printf("Filter: %v, Material: %v", ls.Filter, ls.MaterialSampler.Output)
			concepts.BlendColors(&ls.Filter, &ls.MaterialSampler.Output, 1)
		}

		ls.prevDistSq = ls.hitDistSq
		depth++
		if depth > constants.MaxPortals { // Avoid infinite looping.
			//	dbg := fmt.Sprintf("lightVisible traversed max sectors (p: %v, light: %v)", p, lightBody.Entity)
			//	ls.Player.Notices.Push(dbg)
			return false
		}
		ls.Visited = append(ls.Visited, sector)
		sector = next
	}
	// Some kind of an edge case
	if lightBody.SectorEntity != 0 && ls.Visited[len(ls.Visited)-1].Entity != lightBody.SectorEntity {
		if LogDebug && LogDebugLightHash == ls.Hash && LogDebugLightEntity == lightBody.Entity {
			log.Printf("No intersections, but ended up in a different sector %v than the light sector %v!", ls.Visited[len(ls.Visited)-1].Entity, lightBody.SectorEntity)
		}
		return false
	}

	lightPos := &lightBody.Pos.Render
	// TODO: Use quadtree here, to avoid thrashing memory with the Visited slice
	// Generate entity shadows last. That way if the light is blocked by sector
	// walls, we don't waste time checking/blending lots of bodies or internal
	// segments.
	for _, sector := range ls.Visited {
		for _, seg := range sector.InternalSegments {
			if &seg.Segment == ls.Segment {
				continue
			}
			// Find the intersection with this segment.
			ok := seg.Intersect3D(p, lightPos, &ls.Hit)
			if !ok || ls.Hit[2] < seg.Bottom || ls.Hit[2] > seg.Top {
				// log.Printf("No intersection for internal seg %v|%v\n", seg.A.StringHuman(), seg.B.StringHuman())
				continue // No intersection, skip it!
			}

			ls.Initialize(seg.Surface.Material, seg.Surface.ExtraStages)
			ls.NU = ls.Hit.To2D().Dist(seg.A) / seg.Length
			ls.NV = (seg.Top - ls.Hit[2]) / (seg.Top - seg.Bottom)
			ls.U = ls.NU
			ls.V = ls.NV
			ls.SampleMaterial(seg.Surface.ExtraStages)
			if lit := materials.GetLit(seg.Surface.Material); lit != nil {
				lit.Apply(&ls.MaterialSampler.Output, nil)
			}
			if ls.MaterialSampler.Output[3] >= 0.99 {
				return false
			}
			ls.MaterialSampler.Output[0] = 0
			ls.MaterialSampler.Output[1] = 0
			ls.MaterialSampler.Output[2] = 0
			concepts.BlendColors(&ls.Filter, &ls.MaterialSampler.Output, 1)
		}
		for _, b := range sector.Bodies {
			if !b.IsActive() ||
				b.Entity == lightBody.Entity ||
				b.Entity == ls.InputBody {
				continue
			}
			vis := materials.GetVisible(b.Entity)
			if vis == nil || !vis.IsActive() || vis.Shadow == materials.ShadowNone {
				continue
			}
			switch vis.Shadow {
			case materials.ShadowImage:
				if ok := ls.InitializeRayBody(p, lightPos, b); ok {
					ls.SampleMaterial(nil)
					if ls.MaterialSampler.Output[3]*vis.Opacity > 0.5 {
						return false
					}
				}
			case materials.ShadowSphere:
				if concepts.IntersectLineSphere(p, lightPos, &b.Pos.Render, b.Size.Render[0]*0.5) {
					return false
				}
			case materials.ShadowAABB:
				ext := &concepts.Vector3{b.Size.Render[0], b.Size.Render[0], b.Size.Render[1]}
				if concepts.IntersectLineAABB(p, lightPos, &b.Pos.Render, ext) {
					return false
				}
			}
		}
	}
	if LogDebug && LogDebugLightHash == ls.Hash && LogDebugLightEntity == lightBody.Entity {
		log.Printf("Lit!\n")
	}

	return true
}

var LightSamplerLightsTested, LightSamplerCalcs atomic.Uint64

func (ls *LightSampler) Calculate(world *concepts.Vector3) *concepts.Vector3 {
	ls.Output[0] = 0
	ls.Output[1] = 0
	ls.Output[2] = 0

	LightSamplerCalcs.Add(1)
	lightsTested := 0
	core.QuadTree.Root.RangeClosest(world, true, func(body *core.Body) bool {
		if !body.IsActive() {
			return true
		}
		light := core.GetLight(body.Entity)
		if light == nil || !light.IsActive() {
			return true
		}
		if lightsTested > 100 {
			return false
		}
		lightsTested++
		ls.LightWorld[2] = body.Pos.Render[2] - world[2]
		ls.LightWorld[1] = body.Pos.Render[1] - world[1]
		ls.LightWorld[0] = body.Pos.Render[0] - world[0]
		if LogDebug && LogDebugLightHash == ls.Hash && LogDebugLightEntity == body.Entity {
			log.Printf("Body: %v, World: %v, LightWorld: %v", body.Pos.Render.String(), world.String(), ls.LightWorld.String())
		}
		ls.maxDistSq = ls.LightWorld.Length2()
		ls.maxDist = -1 // Only calculate when necessary
		ls.Filter[3] = 0

		if ls.Normal.Dot(&ls.LightWorld) < 0 {
			return true
		}
		diffuseLight := 1.0
		attenuation := 1.0
		// Is the point right next to the light? Visible by definition.
		if ls.maxDistSq > body.Size.Render[0]*body.Size.Render[0]*0.25 {
			ls.Filter[0] = 0
			ls.Filter[1] = 0
			ls.Filter[2] = 0
			LightSamplerLightsTested.Add(1)
			if !ls.lightVisible(world, body) {
				//log.Printf("Shadowed: %v\n", world.StringHuman())
				return true
			}

			// Calculate light strength.
			if light.Attenuation > 0.0 {
				//log.Printf("%v\n", dist)
				if ls.maxDist < 0 {
					ls.maxDist = math.Sqrt(ls.maxDistSq)
				}
				attenuation = light.Strength / math.Pow(ls.maxDist*2/body.Size.Render[0]+1.0, light.Attenuation)
				//attenuation = 100.0 / dist
			}
			// If it's too far away/dark, ignore it.
			if attenuation < constants.LightAttenuationEpsilon {
				//log.Printf("Too far: %v\n", world.StringHuman())
				return true
			}
		}

		if ls.InputBody != 0 {
			diffuseLight = attenuation
		} else {
			// Normalize
			ls.LightWorld.MulSelf(1.0 / math.Sqrt(ls.maxDistSq))
			diffuseLight = ls.Normal.Dot(&ls.LightWorld) * attenuation
		}
		if ls.Filter[3] == 0 {
			ls.Output[0] += light.Diffuse[0] * diffuseLight
			ls.Output[1] += light.Diffuse[1] * diffuseLight
			ls.Output[2] += light.Diffuse[2] * diffuseLight
		} else {
			a := 1.0 - ls.Filter[3]
			ls.Output[0] += light.Diffuse[0]*diffuseLight*a + ls.Filter[0]
			ls.Output[1] += light.Diffuse[1]*diffuseLight*a + ls.Filter[1]
			ls.Output[2] += light.Diffuse[2]*diffuseLight*a + ls.Filter[2]
		}
		return true
	})
	if LogDebug && LogDebugLightHash == ls.Hash {
		log.Printf("Lightmap value fresh: %v\n", ls.Output.StringHuman(2))
	}
	return &ls.Output
}
