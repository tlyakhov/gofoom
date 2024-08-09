// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package state

import (
	"fmt"
	"log"
	"math"

	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
)

type LightElementType int

//go:generate go run github.com/dmarkham/enumer -type=LightElementType -json
const (
	LightElementFloor LightElementType = iota
	LightElementCeil
	LightElementWall
	LightElementBody
)

// All the data and state required to retrieve/calculate a lightmap texel
type LightElement struct {
	*Config
	Type         LightElementType
	MapIndex     uint64
	Delta        concepts.Vector3
	Output       concepts.Vector3
	Filter       concepts.Vector4
	Intersection concepts.Vector3
	Q            concepts.Vector3
	LightWorld   concepts.Vector3
	InputBody    concepts.Entity
	XorSeed      uint64

	Sector  *core.Sector
	Segment *core.Segment
	Normal  concepts.Vector3
}

func (le *LightElement) Debug() *concepts.Vector3 {
	le.LightmapAddressToWorld(le.Sector, &le.Q, le.MapIndex)
	dbg := le.Q.Mul(1.0 / 64.0)
	le.Output[0] = dbg[0] - math.Floor(dbg[0])
	le.Output[1] = dbg[1] - math.Floor(dbg[1])
	le.Output[2] = dbg[2] - math.Floor(dbg[2])
	return &le.Output
}

func (le *LightElement) Get() *concepts.Vector3 {
	//return le.Debug()
	r := concepts.RngXorShift64(le.XorSeed)
	le.XorSeed = r

	if lmResult, exists := le.Sector.Lightmap.Load(le.MapIndex); exists {
		if uint64(lmResult[3])+constants.MaxLightmapAge >= le.Config.Frame ||
			r%constants.LightmapRefreshDither > 0 {
			le.Output[0] = lmResult[0]
			le.Output[1] = lmResult[1]
			le.Output[2] = lmResult[2]
			return &le.Output
		}
	}
	le.LightmapAddressToWorld(le.Sector, &le.Q, le.MapIndex)
	// Ensure our quantized world location is within Z bounds to avoid
	// weird shadowing.
	fz, cz := le.Sector.PointZ(concepts.DynamicRender, le.Q.To2D())
	if le.Q[2] < fz {
		le.Q[2] = fz
	}
	if le.Q[2] > cz {
		le.Q[2] = cz
	}
	le.Calculate(&le.Q)
	le.Sector.Lightmap.Store(le.MapIndex, concepts.Vector4{
		le.Output[0],
		le.Output[1],
		le.Output[2],
		float64(le.Config.Frame + uint64(r)%constants.LightmapRefreshDither),
	})
	return &le.Output

}

// lightVisible determines whether a given light is visible from a world location.
func (le *LightElement) lightVisible(p *concepts.Vector3, body *core.Body) bool {
	// Always check the starting sector
	if le.lightVisibleFromSector(p, body, le.Sector) {
		return true
	}

	// Check adjacent sectors - this is valuable for edge cases where our voxel
	// is near a boundary
	for _, seg := range le.Sector.Segments {
		if seg.AdjacentSector == 0 || seg.AdjacentSegment == nil || seg.PortalHasMaterial {
			continue
		}
		d2 := seg.AdjacentSegment.DistanceToPoint2(p.To2D())
		if d2 >= le.LightGrid*le.LightGrid {
			continue
		}

		floorZ, ceilZ := seg.AdjacentSegment.Sector.PointZ(concepts.DynamicRender, p.To2D())
		if p[2]-ceilZ > le.LightGrid || floorZ-p[2] > le.LightGrid {
			continue
		}
		if le.lightVisibleFromSector(p, body, seg.AdjacentSegment.Sector) {
			return true
		}
	}

	return false
}

var debugSectorName string = "Starting" // "be4bqmfvn27mek306btg"

// TODO: Make more general
// lightVisibleFromSector determines whether a given light is visible from a world location.
func (le *LightElement) lightVisibleFromSector(p *concepts.Vector3, lightBody *core.Body, sector *core.Sector) bool {
	lightPos := lightBody.Pos.Render

	debugLighting := false
	if constants.DebugLighting {
		dbgName := concepts.NamedFromDb(le.DB, sector.Entity).Name
		debugWallCheck := le.Type == LightElementWall
		debugLighting = constants.DebugLighting && debugWallCheck && dbgName == debugSectorName
	}

	if debugLighting {
		log.Printf("lightVisible: world=%v, light=%v\n", p.StringHuman(), lightPos)
	}
	le.Delta[0] = lightPos[0]
	le.Delta[1] = lightPos[1]
	le.Delta[2] = lightPos[2]
	le.Delta.SubSelf(p)
	maxDist2 := le.Delta.Length2()
	// Is the point right next to the light? Visible by definition.
	if maxDist2 <= lightBody.Size.Render[0]*lightBody.Size.Render[0]*0.25 {
		return true
	}

	// The outer loop traverses portals starting from the sector our target point is in,
	// and finishes in the sector our light is in (unless occluded)
	depth := 0 // We keep track of portaling depth to avoid infinite traversal in weird cases.
	prevDist := -1.0
	for sector != nil {
		if debugLighting {
			log.Printf("Sector: %v\n", sector.Entity)
		}
		// Generate entity shadows
		for _, b := range sector.Bodies {
			if !b.Active || b.Shadow == core.BodyShadowNone ||
				b.Entity == lightBody.Entity ||
				(le.Type == LightElementBody && le.InputBody == b.Entity) {
				continue
			}
			switch b.Shadow {
			case core.BodyShadowSphere:
				if concepts.IntersectLineSphere(p, lightPos, b.Pos.Render, b.Size.Render[0]*0.5) {
					return false
				}
			case core.BodyShadowAABB:
				ext := &concepts.Vector3{b.Size.Render[0], b.Size.Render[0], b.Size.Render[1]}
				if concepts.IntersectLineAABB(p, lightPos, b.Pos.Render, ext) {
					return false
				}
			}
		}
		for _, seg := range sector.InternalSegments {
			if &seg.Segment == le.Segment {
				continue
			}
			// Find the intersection with this segment.
			ok := seg.Intersect3D(p, lightPos, &le.Intersection)
			if !ok || le.Intersection[2] < seg.Bottom || le.Intersection[2] > seg.Top {
				if debugLighting {
					log.Printf("No intersection for internal seg %v|%v\n", seg.A.StringHuman(), seg.B.StringHuman())
				}
				continue // No intersection, skip it!
			}

			sampler := &MaterialSampler{
				Config: le.Config,
				ScaleW: 1024,
				ScaleH: 1024}
			sampler.Initialize(seg.Surface.Material, seg.Surface.ExtraStages)
			sampler.NU = le.Intersection.To2D().Dist(seg.A) / seg.Length
			sampler.NV = (seg.Top - le.Intersection[2]) / (seg.Top - seg.Bottom)
			sampler.U = sampler.NU
			sampler.V = sampler.NV
			sampler.SampleMaterial(seg.Surface.ExtraStages)
			if lit := materials.LitFromDb(seg.DB, seg.Surface.Material); lit != nil {
				lit.Apply(&sampler.Output, nil)
			}
			if sampler.Output[3] >= 0.99 {
				return false
			}
		}

		// Does intersecting the ceiling/floor help us?
		/*denom := sector.CeilNormal.Dot(&le.Delta)
		if denom != 0 {
			planeRayDelta := concepts.Vector3{
				sector.Segments[0].P[0] - p[0],
				sector.Segments[0].P[1] - p[1],
				*sector.TopZ.Render - p[2]}
			t := planeRayDelta.Dot(&sector.CeilNormal) / denom
			if t > 0 {
				//				le.Intersection[2] = le.Delta[2] * t
				le.Intersection[1] = le.Delta[1] * t
				le.Intersection[0] = le.Delta[0] * t
				if sector.IsPointInside2D(le.Intersection.To2D()) {
					return false
				}
			}
		}*/

		var next *core.Sector
		// Since our sectors can be concave, we can't just go through the first portal we find,
		// we have to go through the NEAREST one. Use dist2 to keep track...
		dist2 := maxDist2
		for _, seg := range sector.Segments {
			// Don't occlude the world location with the segment it's located on
			// Segment facing backwards from our ray? skip it.
			if (le.Type == LightElementWall && &seg.Segment == le.Segment) ||
				le.Delta[0]*seg.Normal[0]+le.Delta[1]*seg.Normal[1] > 0 {
				if debugLighting {
					log.Printf("Ignoring segment [or behind] for seg %v|%v\n", seg.P.StringHuman(), seg.Next.P.StringHuman())
				}
				continue
			}

			// Find the intersection with this segment.
			ok := seg.Intersect3D(p, lightPos, &le.Intersection)
			if !ok {
				if debugLighting {
					log.Printf("No intersection for seg %v|%v\n", seg.P.StringHuman(), seg.Next.P.StringHuman())
				}
				continue // No intersection, skip it!
			}

			if debugLighting {
				log.Printf("Intersection for seg %v|%v = %v\n", seg.P.StringHuman(), seg.Next.P.StringHuman(), le.Intersection.StringHuman())
			}

			if seg.AdjacentSector == 0 {
				if debugLighting {
					log.Printf("Occluded behind wall seg %v|%v\n", seg.P.StringHuman(), seg.Next.P.StringHuman())
				}
				return false // This is a wall, that means the light is occluded for sure.
			}

			// Here, we know we have an intersected portal segment. It could still be occluding the light though, since the
			// bottom/top portions could be in the way.
			floorZ, ceilZ := sector.PointZ(concepts.DynamicRender, le.Intersection.To2D())
			floorZ2, ceilZ2 := seg.AdjacentSegment.Sector.PointZ(concepts.DynamicRender, le.Intersection.To2D())
			if debugLighting {
				log.Printf("floorZ: %v, ceilZ: %v, floorZ2: %v, ceilZ2: %v\n", floorZ, ceilZ, floorZ2, ceilZ2)
			}
			if le.Intersection[2] < floorZ2 || le.Intersection[2] > ceilZ2 ||
				le.Intersection[2] < floorZ || le.Intersection[2] > ceilZ {
				if debugLighting {
					log.Printf("Occluded by floor/ceiling gap: %v - %v\n", seg.P.StringHuman(), seg.Next.P.StringHuman())
				}
				return false // Same as wall, we're occluded.
			}

			// If the portal has a transparent material, we need to filter the light
			if seg.PortalHasMaterial {
				sampler := &MaterialSampler{
					Config: le.Config,
					ScaleW: 1024,
					ScaleH: 1024}
				sampler.Initialize(seg.Surface.Material, seg.Surface.ExtraStages)
				sampler.NU = le.Intersection.To2D().Dist(&seg.P) / seg.Length
				sampler.NV = (ceilZ - le.Intersection[2]) / (ceilZ - floorZ)
				sampler.U = sampler.NU
				sampler.V = sampler.NV
				sampler.SampleMaterial(seg.Surface.ExtraStages)
				if lit := materials.LitFromDb(seg.DB, seg.Surface.Material); lit != nil {
					lit.Apply(&sampler.Output, nil)
				}
				if sampler.Output[3] >= 0.99 {
					return false
				}
				//concepts.AsmVector4AddPreMulColorSelf((*[4]float64)(&le.Filter), (*[4]float64)(&le.Material))
				le.Filter.AddPreMulColorSelf(&sampler.Output)
			}

			// Get the square of the distance to the intersection (from the target point)
			idist2 := le.Intersection.Dist2(p)

			// If the difference between the intersected distance and the light distance is
			// within the bounding radius of our light, our light is right on a portal boundary and visible.
			if math.Abs(idist2-maxDist2) < lightBody.Size.Render[0]*0.5 {
				return true
			}

			if idist2-dist2 > constants.IntersectEpsilon {
				if debugLighting {
					log.Printf("Found intersection point farther than one we've already discovered for this sector: %v > %v\n", idist2, dist2)
				}
				// If the current intersection point is farther than one we already have for this sector, we have a concavity. Keep looking.
				continue
			}

			if prevDist-idist2 > constants.IntersectEpsilon {
				if debugLighting {
					log.Printf("Found intersection point before the previous sector: %v < %v\n", idist2, prevDist)
				}
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
			dbg := fmt.Sprintf("lightVisible traversed max sectors (p: %v, light: %v)", p, lightBody.Entity)
			le.DebugNotices.Push(dbg)
			return false
		}
		if next == nil && lightBody.SectorEntity != 0 && sector.Entity != lightBody.SectorEntity {
			if debugLighting {
				log.Printf("No intersections, but ended up in a different sector than the light!\n")
			}
			return false
		}
		sector = next
	}

	if debugLighting {
		log.Printf("Lit!\n")
	}
	return true
}

func (le *LightElement) Calculate(world *concepts.Vector3) *concepts.Vector3 {
	le.Output[0] = 0
	le.Output[1] = 0
	le.Output[2] = 0

	/*refs := make([]concepts.Entity, 0)
	for _, entity := range le.Sector.PVL {
		refs = append(refs, entity)
	}
	sort.SliceStable(refs, func(i, j int) bool {
		return refs[i].Entity < refs[j].Entity
	})*/

	for entity, body := range le.Sector.PVL {
		light := core.LightFromDb(body.DB, entity)
		if !light.IsActive() {
			continue
		}

		if !body.IsActive() {
			continue
		}
		le.Filter[0] = 0
		le.Filter[1] = 0
		le.Filter[2] = 0
		le.Filter[3] = 0
		le.LightWorld[0] = body.Pos.Render[0]
		le.LightWorld[1] = body.Pos.Render[1]
		le.LightWorld[2] = body.Pos.Render[2]
		le.LightWorld.SubSelf(world)
		dist := le.LightWorld.Length()
		diffuseLight := 1.0
		attenuation := 1.0
		if dist != 0 {
			// Normalize
			le.LightWorld.MulSelf(1.0 / dist)
			// Calculate light strength.
			if light.Attenuation > 0.0 {
				//log.Printf("%v\n", dist)
				attenuation = light.Strength / math.Pow(dist*2/body.Size.Render[0]+1.0, light.Attenuation)
				//attenuation = 100.0 / dist
			}
			// If it's too far away/dark, ignore it.
			if attenuation < constants.LightAttenuationEpsilon {
				//log.Printf("Too far: %v\n", world.StringHuman())
				continue
			}
			if !le.lightVisible(world, body) {
				//log.Printf("Shadowed: %v\n", world.StringHuman())
				continue
			}
		}
		if le.Type == LightElementBody {
			diffuseLight = attenuation
		} else {
			diffuseLight = le.Normal.Dot(&le.LightWorld) * attenuation
		}

		if diffuseLight < 0 {
			//le.Output[0] = 1
			continue
		}
		if le.Filter[3] == 0 {
			le.Output[0] += light.Diffuse[0] * diffuseLight
			le.Output[1] += light.Diffuse[1] * diffuseLight
			le.Output[2] += light.Diffuse[2] * diffuseLight
		} else {
			a := 1.0 - le.Filter[3]
			le.Output[0] += light.Diffuse[0]*diffuseLight*a + le.Filter[0]
			le.Output[1] += light.Diffuse[1]*diffuseLight*a + le.Filter[1]
			le.Output[2] += light.Diffuse[2]*diffuseLight*a + le.Filter[2]
		}
	}
	return &le.Output
}
