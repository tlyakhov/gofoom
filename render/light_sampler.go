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
	"tlyakhov/gofoom/dynamic"
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
	Normal  concepts.Vector3

	prevDistSq, hitDistSq, maxDist2 float64

	xorSeed uint64
	tree    *core.Quadtree
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
	fz, cz := ls.Sector.ZAt(dynamic.Render, ls.Q.To2D())
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

	// Always check the starting sector
	if ls.lightVisibleFromSector(p, body, ls.Sector) {
		return true
	}

	// Check adjacent sectors - this is valuable for edge cases where our voxel
	// is near a boundary
	for _, seg := range ls.Sector.Segments {
		if seg.AdjacentSector == 0 || seg.AdjacentSegment == nil || seg.PortalHasMaterial {
			continue
		}
		d2 := seg.AdjacentSegment.DistanceToPoint2(p.To2D())
		if d2 >= ls.LightGrid*ls.LightGrid*2 {
			continue
		}

		floorZ, ceilZ := seg.AdjacentSegment.Sector.ZAt(dynamic.Render, p.To2D())
		if p[2]-ceilZ > ls.LightGrid || floorZ-p[2] > ls.LightGrid {
			continue
		}
		if ls.lightVisibleFromSector(p, body, seg.AdjacentSegment.Sector) {
			return true
		}
	}

	return false
}

func (ls *LightSampler) intersect(sector *core.Sector, p *concepts.Vector3, lightBody *core.Body, peekIntoInner bool) (hitSegment *core.SectorSegment, next *core.Sector) {
	lightPos := &lightBody.Pos.Render

	// Help the compiler out by pre-defining all the local stuff in the outer
	// scope. I wonder how much performance cost we have with this stuff on the
	// stack (64 bytes?)
	var i2d *concepts.Vector2
	var adj *core.Sector
	var floorZ, floorZ2, ceilZ, ceilZ2, intersectionDistSq float64
	for _, seg := range sector.Segments {
		if LogDebug && LogDebugLightHash == ls.Hash && LogDebugLightEntity == lightBody.Entity {
			log.Printf("    Checking segment [%v]-[%v]\n", seg.P.StringHuman(), seg.Next.P.StringHuman())
		}
		// Don't occlude the world location with the segment it's located on
		if &seg.Segment == ls.Segment {
			continue
		}
		// When peeking into inner sectors, ignore their portals. We only care
		// about the light ray _entering_ the inner sector.
		if peekIntoInner && seg.AdjacentSector != 0 {
			if LogDebug && LogDebugLightHash == ls.Hash && LogDebugLightEntity == lightBody.Entity {
				log.Printf("    Ignoring portal to %v in inner sector.", seg.AdjacentSector)
			}
			continue
		}
		/*		inner	  n>0  result
				   	0		1		1
					0		0		0
					1		0		1
					1		1		0
		*/
		normalFacing := (ls.LightWorld[0]*seg.Normal[0]+ls.LightWorld[1]*seg.Normal[1] > 0)
		if peekIntoInner != normalFacing {
			if LogDebug && LogDebugLightHash == ls.Hash && LogDebugLightEntity == lightBody.Entity {
				log.Printf("    Ignoring segment [or behind]. inner? = %v, normal? = %v. Dot: %v", peekIntoInner, normalFacing, ls.LightWorld[0]*seg.Normal[0]+ls.LightWorld[1]*seg.Normal[1])
			}
			continue
		}

		// Find the intersection with this segment.
		if !seg.Intersect3D(p, lightPos, &ls.IntersectionTest) {
			if LogDebug && LogDebugLightHash == ls.Hash && LogDebugLightEntity == lightBody.Entity {
				log.Printf("    No intersection\n")
			}
			continue // No intersection, skip it!
		}

		if LogDebug && LogDebugLightHash == ls.Hash && LogDebugLightEntity == lightBody.Entity {
			log.Printf("    Intersection = [%v]\n", ls.IntersectionTest.StringHuman(2))
		}

		nudgeExitingRay := 0.0
		// This segment could either be:
		if seg.AdjacentSector != 0 {
			// A portal!
			adj = seg.AdjacentSegment.Sector
			if LogDebug && LogDebugLightHash == ls.Hash && LogDebugLightEntity == lightBody.Entity {
				log.Printf("    Portal to %v", adj.Entity)
			}
		} else if peekIntoInner {
			// An inner segment!
			adj = sector
			if LogDebug && LogDebugLightHash == ls.Hash && LogDebugLightEntity == lightBody.Entity {
				log.Printf("    Inner to %v", adj.Entity)
			}
		} else if !sector.Outer.Empty() {
			// We have a non-portal segment and we're not checking inner sector,
			// but our sector itself is an inner one. Let's go out
			adj = sector.OuterAt(ls.IntersectionTest.To2D())
			// Why? This is to handle the edge case when we have a light ray
			// grazing a corner of an inner sector. If this happens, we need to
			// nudge the intersection distance for the _exiting_ light ray to
			// avoid an infinite loop of intersections, ping-ponging between the
			// inner and outer sector.
			nudgeExitingRay = constants.IntersectEpsilon
			if LogDebug && LogDebugLightHash == ls.Hash && LogDebugLightEntity == lightBody.Entity {
				log.Printf("    Out to %v", adj.Entity)
			}
		} else {
			if LogDebug && LogDebugLightHash == ls.Hash && LogDebugLightEntity == lightBody.Entity {
				log.Printf("    Occluded behind wall seg %v|%v\n", seg.P.StringHuman(), seg.Next.P.StringHuman())
			}
			// A wall!
			return seg, nil // This is a wall, that means the light is occluded for sure.
		}

		// Here, we know we have an intersected portal segment. It could still be occluding the light though, since the
		// bottom/top portions could be in the way.
		i2d = ls.IntersectionTest.To2D()
		if ls.IntersectionTest[2] < sector.Min[2] || ls.IntersectionTest[2] > sector.Max[2] {
			if LogDebug && LogDebugLightHash == ls.Hash && LogDebugLightEntity == lightBody.Entity {
				log.Printf("    Occluded by sector min/max %v - %v\n", seg.P.StringHuman(), seg.Next.P.StringHuman())
			}
			return seg, nil // Same as wall, we're occluded.
		}
		floorZ, ceilZ = sector.ZAt(dynamic.Render, i2d)
		// log.Printf("floorZ: %v, ceilZ: %v, floorZ2: %v, ceilZ2: %v\n", floorZ, ceilZ, floorZ2, ceilZ2)
		if ls.IntersectionTest[2] < floorZ || ls.IntersectionTest[2] > ceilZ {
			if LogDebug && LogDebugLightHash == ls.Hash && LogDebugLightEntity == lightBody.Entity {
				log.Printf("    Occluded by floor/ceiling gap: %v - %v\n", seg.P.StringHuman(), seg.Next.P.StringHuman())
			}
			return seg, nil // Same as wall, we're occluded.
		}
		if !peekIntoInner {
			floorZ2, ceilZ2 = adj.ZAt(dynamic.Render, i2d)
			// log.Printf("floorZ: %v, ceilZ: %v, floorZ2: %v, ceilZ2: %v\n", floorZ, ceilZ, floorZ2, ceilZ2)
			if ls.IntersectionTest[2] < floorZ2 || ls.IntersectionTest[2] > ceilZ2 {
				if LogDebug && LogDebugLightHash == ls.Hash && LogDebugLightEntity == lightBody.Entity {
					log.Printf("    Occluded by floor/ceiling gap: %v - %v\n", seg.P.StringHuman(), seg.Next.P.StringHuman())
				}
				return seg, nil // Same as wall, we're occluded.
			}
		}

		// Get the square of the distance to the intersection (from the target point)
		intersectionDistSq = ls.IntersectionTest.Dist2(p) + nudgeExitingRay

		// If the difference between the intersected distance and the light distance is
		// within the bounding radius of our light, our light is right on a portal boundary and visible.
		if math.Abs(intersectionDistSq-ls.maxDist2) < lightBody.Size.Render[0]*0.5 {
			if LogDebug && LogDebugLightHash == ls.Hash && LogDebugLightEntity == lightBody.Entity {
				log.Printf("    Light is within size on portal boundary. Intersection: %v vs. Max Dist %v", math.Sqrt(intersectionDistSq), math.Sqrt(ls.maxDist2))
			}
			return nil, nil
		}

		if intersectionDistSq-ls.hitDistSq >= 0 {
			if LogDebug && LogDebugLightHash == ls.Hash && LogDebugLightEntity == lightBody.Entity {
				log.Printf("    Found intersection point farther than one we've already discovered for this sector: %v > %v\n", math.Sqrt(intersectionDistSq), math.Sqrt(ls.hitDistSq))
			}
			// If the current intersection point is farther than one we already have for this sector, we have a concavity. Keep looking.
			continue
		}

		if ls.prevDistSq-intersectionDistSq > 0 {
			if LogDebug && LogDebugLightHash == ls.Hash && LogDebugLightEntity == lightBody.Entity {
				log.Printf("    Found intersection point before the previous sector: %v <= %v\n", math.Sqrt(intersectionDistSq), math.Sqrt(ls.prevDistSq))
			}
			// If the current intersection point is BEHIND the last one, we went backwards?
			continue
		}

		if LogDebug && LogDebugLightHash == ls.Hash && LogDebugLightEntity == lightBody.Entity {
			log.Printf("    Found a portal to %v without impediment at dist %v.\n", adj.Entity, math.Sqrt(intersectionDistSq))
		}
		// We're in the clear! We have a portal (or boundary between inner/outer
		// sector) without anything impending the light.
		ls.hitDistSq = intersectionDistSq
		ls.Hit = ls.IntersectionTest
		hitSegment = seg
		next = adj
	}
	return
}

// lightVisibleFromSector determines whether a given light is visible from a world location.
func (ls *LightSampler) lightVisibleFromSector(p *concepts.Vector3, lightBody *core.Body, sector *core.Sector) bool {
	if LogDebug && LogDebugLightHash == ls.Hash && LogDebugLightEntity == lightBody.Entity {
		log.Printf("LightSampler.lightVisibleFromSector: START")
		log.Printf("world=%v, light=%v\n", p.String(), lightBody.Pos.Render.String())
	}
	var next *core.Sector
	var hitSegment *core.SectorSegment

	// The outer loop traverses portals starting from the sector our target point is in,
	// and finishes in the sector our light is in (unless occluded)
	depth := 0 // We keep track of portaling depth to avoid infinite traversal in weird cases.
	ls.Visited = ls.Visited[:0]
	ls.prevDistSq = -1.0
	for sector != nil {
		// Since our sectors can be concave or have inner sectors (holes), we
		// can't just go through the first portal we find, we have to go through
		// the NEAREST one. Use hitDistSq to keep track...
		ls.hitDistSq = ls.maxDist2

		if LogDebug && LogDebugLightHash == ls.Hash && LogDebugLightEntity == lightBody.Entity {
			log.Printf("  Checking sector %v, max dist %v", sector.Entity, math.Sqrt(ls.maxDist2))
		}
		//Intersect this sector
		seg, adj := ls.intersect(sector, p, lightBody, false)
		if seg != nil && adj == nil {
			// Occluded
			return false
		}
		next = adj
		hitSegment = seg

		// Check the inner sectors
		for _, e := range sector.Inner {
			if e == 0 {
				continue
			}
			if LogDebug && LogDebugLightHash == ls.Hash && LogDebugLightEntity == lightBody.Entity {
				log.Printf("  Visiting inner sector %v, max dist %v", e, ls.maxDist2)
			}
			if inner := core.GetSector(e); inner != nil {
				seg, adj = ls.intersect(inner, p, lightBody, true)
				if seg != nil && adj == nil {
					// Occluded
					return false
				}
				if adj != nil {
					next = adj
					hitSegment = seg
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
			floorZ, ceilZ := sector.ZAt(dynamic.Render, i2d)
			ls.Initialize(hitSegment.Surface.Material, hitSegment.Surface.ExtraStages)
			ls.NU = ls.Hit.To2D().Dist(&hitSegment.P) / hitSegment.Length
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
			ls.Filter[3] += ls.MaterialSampler.Output[3]
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
	ls.tree.Root.RangeClosest(world, true, func(body *core.Body) bool {
		if !body.IsActive() {
			return true
		}
		light := core.GetLight(body.Entity)
		if light == nil || !light.IsActive() {
			return true
		}
		if lightsTested > 100 {
			//	return false
		}
		lightsTested++
		LightSamplerLightsTested.Add(1)
		ls.LightWorld[2] = body.Pos.Render[2] - world[2]
		ls.LightWorld[1] = body.Pos.Render[1] - world[1]
		ls.LightWorld[0] = body.Pos.Render[0] - world[0]
		if LogDebug && LogDebugLightHash == ls.Hash && LogDebugLightEntity == body.Entity {
			log.Printf("Body: %v, World: %v, LightWorld: %v", body.Pos.Render.String(), world.String(), ls.LightWorld.String())
		}
		ls.maxDist2 = ls.LightWorld.Length2()
		ls.Filter[3] = 0

		diffuseLight := 1.0
		attenuation := 1.0
		// Is the point right next to the light? Visible by definition.
		if ls.maxDist2 > body.Size.Render[0]*body.Size.Render[0]*0.25 {
			ls.Filter[0] = 0
			ls.Filter[1] = 0
			ls.Filter[2] = 0
			if !ls.lightVisible(world, body) {
				//log.Printf("Shadowed: %v\n", world.StringHuman())
				return true
			}
			// Calculate light strength.
			if light.Attenuation > 0.0 {
				//log.Printf("%v\n", dist)
				dist := math.Sqrt(ls.maxDist2)
				attenuation = light.Strength / math.Pow(dist*2/body.Size.Render[0]+1.0, light.Attenuation)
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
			ls.LightWorld.MulSelf(1.0 / math.Sqrt(ls.maxDist2))
			diffuseLight = ls.Normal.Dot(&ls.LightWorld) * attenuation
		}

		if diffuseLight < 0 {
			//le.Output[0] = 1
			return true
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
