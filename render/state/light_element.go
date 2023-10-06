package state

import (
	"log"
	"math"

	"tlyakhov/gofoom/behaviors"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/core"
)

type LightElement struct {
	Normal   concepts.Vector3
	Sector   *core.PhysicalSector
	Segment  *core.Segment
	MapIndex uint32
	Lightmap []concepts.Vector3
}

func (le *LightElement) Get(wall bool) concepts.Vector3 {
	result := le.Lightmap[le.MapIndex]
	if result.X >= 0 {
		return result
	}

	var q concepts.Vector3
	if !wall {
		q = le.Sector.LightmapAddressToWorld(le.MapIndex, le.Normal.Z > 0)
	} else {
		//log.Printf("Lightmap element doesn't exist: %v, %v, %v\n", le.Sector.ID, le.MapIndex, le.Segment.ID)
		q = le.Segment.LightmapAddressToWorld(le.MapIndex)
	}
	result = le.Calculate(q)
	//dbg := q.Mul(1.0 / 64.0)
	//result = concepts.Vector3{dbg.X - math.Floor(dbg.X), dbg.Y - math.Floor(dbg.Y), dbg.Z - math.Floor(dbg.Z)}

	le.Lightmap[le.MapIndex] = result
	return result
}

// lightVisible determines whether a given light is visible from a world location.
func (le *LightElement) lightVisible(p concepts.Vector3, e *core.PhysicalEntity) bool {
	debugSectorID := "tallBoy" // "be4bqmfvn27mek306btg"
	debugWallCheck := le.Normal.Z != 0
	if constants.DebugLighting && debugWallCheck && le.Sector.ID == debugSectorID {
		log.Printf("lightVisible: world=%v, light=%v\n", p.StringHuman(), e.Pos)
	}
	p2d := p.To2D()
	sector := le.Sector

	floorZ, ceilZ := sector.CalcFloorCeilingZ(p2d)
	if ceilZ-floorZ <= constants.IntersectEpsilon {
		return false
	}

	delta := e.Pos.Sub(p)
	maxDist2 := delta.Length2()
	// Is the point right next to the light? Visible by definition.
	if maxDist2 <= e.BoundingRadius*e.BoundingRadius {
		return true
	}

	// The outer loop traverses portals starting from the sector our target point is in,
	// and finishes in the sector our light is in (unless occluded)
	depth := 0 // We keep track of portaling depth to avoid infinite traversal in weird cases.
	prevDist := -1.0
	for sector != nil {
		if constants.DebugLighting && debugWallCheck && le.Sector.ID == debugSectorID {
			log.Printf("Sector: %v\n", sector.ID)
		}
		var next *core.PhysicalSector
		// Since our sectors can be concave, we can't just go through the first portal we find,
		// we have to go through the NEAREST one. Use dist2 to keep track...
		dist2 := maxDist2
		for _, seg := range sector.Segments {
			// Don't occlude the world location with the segment it's located on
			// Segment facing backwards from our ray? skip it.
			if (le.Normal.Z == 0 && seg == le.Segment) || delta.To2D().Dot(seg.Normal) > 0 {
				if constants.DebugLighting && debugWallCheck && le.Sector.ID == debugSectorID {
					log.Printf("Ignoring segment [or behind] for seg %v|%v\n", seg.P.StringHuman(), seg.Next.P.StringHuman())
				}
				continue
			}

			// Find the intersection with this segment.
			isect, ok := seg.Intersect3D(p, e.Pos)
			if !ok {
				if constants.DebugLighting && debugWallCheck && le.Sector.ID == debugSectorID {
					log.Printf("No intersection for seg %v|%v\n", seg.P.StringHuman(), seg.Next.P.StringHuman())
				}
				continue // No intersection, skip it!
			}

			if constants.DebugLighting && debugWallCheck && le.Sector.ID == debugSectorID {
				log.Printf("Intersection for seg %v|%v = %v\n", seg.P.StringHuman(), seg.Next.P.StringHuman(), isect.StringHuman())
			}

			if seg.AdjacentSector == nil {
				if constants.DebugLighting && debugWallCheck && le.Sector.ID == debugSectorID {
					log.Printf("Occluded behind wall seg %v|%v\n", seg.P.StringHuman(), seg.Next.P.StringHuman())
				}
				return false // This is a wall, that means the light is occluded for sure.
			}

			// Here, we know we have an intersected portal segment. It could still be occluding the light though, since the
			// bottom/top portions could be in the way.
			floorZ, ceilZ := sector.CalcFloorCeilingZ(isect.To2D())
			floorZ2, ceilZ2 := seg.AdjacentSector.Physical().CalcFloorCeilingZ(isect.To2D())
			if constants.DebugLighting && debugWallCheck && le.Sector.ID == debugSectorID {
				log.Printf("floorZ: %v, ceilZ: %v, floorZ2: %v, ceilZ2: %v\n", floorZ, ceilZ, floorZ2, ceilZ2)
			}
			if isect.Z < floorZ2 || isect.Z > ceilZ2 || isect.Z < floorZ || isect.Z > ceilZ {
				if constants.DebugLighting && debugWallCheck && le.Sector.ID == debugSectorID {
					log.Printf("Occluded by floor/ceiling gap: %v - %v\n", seg.P.StringHuman(), seg.Next.P.StringHuman())
				}
				return false // Same as wall, we're occluded.
			}

			// Get the square of the distance to the intersection (from the target point)
			d := isect.Dist2(p)

			if d > dist2 || d <= prevDist {
				if constants.DebugLighting && debugWallCheck && le.Sector.ID == debugSectorID {
					log.Printf("Found intersection point farther than one we've already discovered: %v > %v\n", d, dist2)
				}
				// If the current intersection point is farther than one we already have for this sector,
				// or BEHIND the last one, we have a concavity. Keep looking.
				continue
			}

			// We're in the clear! Move to the next adjacent sector.
			dist2 = d
			next = seg.AdjacentSector.Physical()
		}
		prevDist = dist2
		depth++
		if depth > 100 { // Avoid infinite looping.
			log.Printf("warning: lightVisible traversed > 100 sectors.\n")
			return false
		}
		if next == nil && sector != e.Sector.Physical() {
			if constants.DebugLighting && debugWallCheck && le.Sector.ID == debugSectorID {
				log.Printf("No intersections, but ended up in a different sector than the light!\n")
			}
			return false
		}
		sector = next
	}

	if constants.DebugLighting && debugWallCheck && le.Sector.ID == debugSectorID {
		log.Printf("Lit!\n")
	}
	return true
}

func (le *LightElement) Calculate(world concepts.Vector3) concepts.Vector3 {
	diffuseSum := concepts.Vector3{}

	for _, lightEntity := range le.Sector.PVL {
		if !lightEntity.Physical().Active {
			continue
		}

		for _, b := range lightEntity.Physical().Behaviors {
			lb, ok := b.(*behaviors.Light)
			if !ok {
				continue
			}
			delta := lightEntity.Physical().Pos.Sub(world)
			dist := delta.Length()
			attenuation := 1.0
			if dist != 0 {
				// Normalize
				delta = delta.Mul(1.0 / dist)
				// Calculate light strength.
				if lb.Attenuation > 0.0 {
					//log.Printf("%v\n", dist)
					attenuation = lb.Strength / math.Pow(dist/lightEntity.Physical().BoundingRadius+1.0, lb.Attenuation)
					//attenuation = 100.0 / dist
				}
				// If it's too far away/dark, ignore it.
				if attenuation < constants.LightAttenuationEpsilon {
					//log.Printf("Too far: %v\n", world.StringHuman())
					continue
				}
				if !le.lightVisible(world, lightEntity.Physical()) {
					//log.Printf("Shadowed: %v\n", world.StringHuman())
					continue
				}
			}

			diffuseLight := le.Normal.Dot(delta) * attenuation
			if diffuseLight > 0 {
				diffuseSum = diffuseSum.Add(lb.Diffuse.Mul(diffuseLight))
			}
		}
	}
	return diffuseSum
}
