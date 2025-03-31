// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package render

import (
	"math"
	"sync/atomic"

	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/dynamic"
	"tlyakhov/gofoom/ecs"
)

// All the data and state required to retrieve/calculate a lightmap voxel
type LightSampler struct {
	MaterialSampler
	Hash         uint64
	Output       concepts.Vector3
	Filter       concepts.Vector4
	Intersection concepts.Vector3
	Q            concepts.Vector3
	LightWorld   concepts.Vector3
	InputBody    ecs.Entity
	Visited      []*core.Sector

	Sector  *core.Sector
	Segment *core.Segment
	Normal  concepts.Vector3

	xorSeed  uint64
	maxDist2 float64
	tree     *core.Quadtree
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
		if lmResult.Timestamp+constants.MaxLightmapAge >= ls.Universe.Frame ||
			!concepts.RngDecide(r, constants.LightmapRefreshDither) {
			ls.Output = lmResult.Light
			return &ls.Output
		}
	}
	ls.LightmapHashToWorld(ls.Sector, &ls.Q, ls.Hash)
	// Ensure our quantized world location is within Z bounds to avoid
	// weird shadowing.
	fz, cz := ls.Sector.ZAt(dynamic.DynamicRender, ls.Q.To2D())
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
		Light: concepts.Vector3{
			ls.Output[0],
			ls.Output[1],
			ls.Output[2]},
		Timestamp: ls.Universe.Frame,
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

		floorZ, ceilZ := seg.AdjacentSegment.Sector.ZAt(dynamic.DynamicRender, p.To2D())
		if p[2]-ceilZ > ls.LightGrid || floorZ-p[2] > ls.LightGrid {
			continue
		}
		if ls.lightVisibleFromSector(p, body, seg.AdjacentSegment.Sector) {
			return true
		}
	}

	return false
}

// lightVisibleFromSector determines whether a given light is visible from a world location.
func (ls *LightSampler) lightVisibleFromSector(p *concepts.Vector3, lightBody *core.Body, sector *core.Sector) bool {
	lightPos := lightBody.Pos.Render
	// log.Printf("lightVisible: world=%v, light=%v\n", p.StringHuman(),
	// lightPos)

	// The outer loop traverses portals starting from the sector our target point is in,
	// and finishes in the sector our light is in (unless occluded)
	depth := 0 // We keep track of portaling depth to avoid infinite traversal in weird cases.
	prevDist := -1.0
	ls.Visited = ls.Visited[:0]
	for sector != nil {
		// log.Printf("Sector: %v\n", sector.Entity)
		/*denom := sector.Top.Normal.Dot(&le.Delta)
		if denom != 0 {
			t := (sector.Segments[0].P[0]-p[0])*sector.Top.Normal[0] +
				(sector.Segments[0].P[1]-p[1])*sector.Top.Normal[1] +
				(*sector.Top.Z.Render-p[2])*sector.Top.Normal[2]
			t /= denom
			if t > 0.1 && t*t < maxDist2 {
				return false
			}
		}
		denom = sector.Bottom.Normal.Dot(&le.Delta)
		if denom != 0 {
			t := (sector.Segments[0].P[0]-p[0])*sector.Bottom.Normal[0] +
				(sector.Segments[0].P[1]-p[1])*sector.Bottom.Normal[1] +
				(*sector.Bottom.Z.Render-p[2])*sector.Bottom.Normal[2]
			t /= denom
			if t > 0.1 && t*t < maxDist2 {
				return false
			}
		}*/

		var next *core.Sector
		// Since our sectors can be concave, we can't just go through the first portal we find,
		// we have to go through the NEAREST one. Use dist2 to keep track...
		dist2 := ls.maxDist2
		for _, seg := range sector.Segments {
			// Don't occlude the world location with the segment it's located on
			// Segment facing backwards from our ray? skip it.
			if &seg.Segment == ls.Segment ||
				ls.LightWorld[0]*seg.Normal[0]+ls.LightWorld[1]*seg.Normal[1] > 0 {
				// log.Printf("Ignoring segment [or behind] for seg %v|%v\n", seg.P.StringHuman(), seg.Next.P.StringHuman())
				continue
			}

			// Find the intersection with this segment.
			ok := seg.Intersect3D(p, lightPos, &ls.Intersection)
			if !ok {
				// log.Printf("No intersection for seg %v|%v\n", seg.P.StringHuman(), seg.Next.P.StringHuman())
				continue // No intersection, skip it!
			}

			// log.Printf("Intersection for seg %v|%v = %v\n", seg.P.StringHuman(), seg.Next.P.StringHuman(), le.Intersection.StringHuman())

			if seg.AdjacentSector == 0 {
				// log.Printf("Occluded behind wall seg %v|%v\n", seg.P.StringHuman(), seg.Next.P.StringHuman())
				return false // This is a wall, that means the light is occluded for sure.
			}

			// Here, we know we have an intersected portal segment. It could still be occluding the light though, since the
			// bottom/top portions could be in the way.
			i2d := ls.Intersection.To2D()
			floorZ, ceilZ := sector.ZAt(dynamic.DynamicRender, i2d)
			floorZ2, ceilZ2 := seg.AdjacentSegment.Sector.ZAt(dynamic.DynamicRender, i2d)
			// log.Printf("floorZ: %v, ceilZ: %v, floorZ2: %v, ceilZ2: %v\n", floorZ, ceilZ, floorZ2, ceilZ2)
			if ls.Intersection[2] < floorZ2 || ls.Intersection[2] > ceilZ2 ||
				ls.Intersection[2] < floorZ || ls.Intersection[2] > ceilZ {
				// log.Printf("Occluded by floor/ceiling gap: %v - %v\n", seg.P.StringHuman(), seg.Next.P.StringHuman())
				return false // Same as wall, we're occluded.
			}

			// If the portal has a transparent material, we need to filter the light
			if seg.PortalHasMaterial {
				ls.Initialize(seg.Surface.Material, seg.Surface.ExtraStages)
				ls.NU = ls.Intersection.To2D().Dist(&seg.P) / seg.Length
				ls.NV = (ceilZ - ls.Intersection[2]) / (ceilZ - floorZ)
				ls.U = ls.NU
				ls.V = ls.NV
				ls.SampleMaterial(seg.Surface.ExtraStages)
				if lit := materials.GetLit(seg.Universe, seg.Surface.Material); lit != nil {
					lit.Apply(&ls.MaterialSampler.Output, nil)
				}
				if ls.MaterialSampler.Output[3] >= 0.99 {
					return false
				}
				concepts.BlendColors(&ls.Filter, &ls.MaterialSampler.Output, 1)
			}

			// Get the square of the distance to the intersection (from the target point)
			idist2 := ls.Intersection.Dist2(p)

			// If the difference between the intersected distance and the light distance is
			// within the bounding radius of our light, our light is right on a portal boundary and visible.
			if math.Abs(idist2-ls.maxDist2) < lightBody.Size.Render[0]*0.5 {
				return true
			}

			if idist2-dist2 > constants.IntersectEpsilon {
				// log.Printf("Found intersection point farther than one we've already discovered for this sector: %v > %v\n", idist2, dist2)
				// If the current intersection point is farther than one we already have for this sector, we have a concavity. Keep looking.
				continue
			}

			if prevDist-idist2 > constants.IntersectEpsilon {
				// log.Printf("Found intersection point before the previous sector: %v < %v\n", idist2, prevDist)
				// If the current intersection point is BEHIND the last one, we went backwards?
				continue
			}

			// We're in the clear! Move to the next adjacent sector.
			dist2 = idist2
			next = seg.AdjacentSegment.Sector
		}
		prevDist = dist2
		depth++
		if depth > constants.MaxPortals { // Avoid infinite looping.
			//	dbg := fmt.Sprintf("lightVisible traversed max sectors (p: %v, light: %v)", p, lightBody.Entity)
			//	ls.Player.Notices.Push(dbg)
			return false
		}
		if next == nil && lightBody.SectorEntity != 0 && sector.Entity != lightBody.SectorEntity {
			// log.Printf("No intersections, but ended up in a different sector than the light!\n")
			return false
		}
		ls.Visited = append(ls.Visited, sector)
		sector = next
	}

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
			ok := seg.Intersect3D(p, lightPos, &ls.Intersection)
			if !ok || ls.Intersection[2] < seg.Bottom || ls.Intersection[2] > seg.Top {
				// log.Printf("No intersection for internal seg %v|%v\n", seg.A.StringHuman(), seg.B.StringHuman())
				continue // No intersection, skip it!
			}

			ls.Initialize(seg.Surface.Material, seg.Surface.ExtraStages)
			ls.NU = ls.Intersection.To2D().Dist(seg.A) / seg.Length
			ls.NV = (seg.Top - ls.Intersection[2]) / (seg.Top - seg.Bottom)
			ls.U = ls.NU
			ls.V = ls.NV
			ls.SampleMaterial(seg.Surface.ExtraStages)
			if lit := materials.GetLit(seg.Universe, seg.Surface.Material); lit != nil {
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
			vis := materials.GetVisible(ls.Universe, b.Entity)
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
				if concepts.IntersectLineSphere(p, lightPos, b.Pos.Render, b.Size.Render[0]*0.5) {
					return false
				}
			case materials.ShadowAABB:
				ext := &concepts.Vector3{b.Size.Render[0], b.Size.Render[0], b.Size.Render[1]}
				if concepts.IntersectLineAABB(p, lightPos, b.Pos.Render, ext) {
					return false
				}
			}
		}
	}

	// log.Printf("Lit!\n")
	return true
}

var LightSamplerLightsTested, LightSamplerCalcs atomic.Uint64

func (ls *LightSampler) Calculate(world *concepts.Vector3) *concepts.Vector3 {
	ls.Output[0] = 0
	ls.Output[1] = 0
	ls.Output[2] = 0

	LightSamplerCalcs.Add(1)
	ls.tree.Root.RangePlane(world, &ls.Normal, true, func(body *core.Body) bool {
		LightSamplerLightsTested.Add(1)
		if !body.IsActive() {
			return true
		}
		light := core.GetLight(ls.Universe, body.Entity)
		if light == nil || !light.IsActive() {
			return true
		}
		ls.LightWorld[2] = body.Pos.Render[2] - world[2]
		ls.LightWorld[1] = body.Pos.Render[1] - world[1]
		ls.LightWorld[0] = body.Pos.Render[0] - world[0]
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
	return &ls.Output
}
