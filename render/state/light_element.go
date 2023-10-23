package state

import (
	"fmt"
	"log"
	"math"

	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
)

type LightElement struct {
	*Slice
	MapIndex uint32
}

func (le *LightElement) Debug(wall bool) *concepts.Vector3 {
	var q = new(concepts.Vector3)
	if !wall {
		le.Sector.LightmapAddressToWorld(q, le.MapIndex, le.Normal[2] > 0)
	} else {
		//log.Printf("Lightmap element doesn't exist: %v, %v, %v\n", le.Sector.Name, le.MapIndex, le.Segment.Name)
		le.Segment.LightmapAddressToWorld(q, le.MapIndex)
	}
	dbg := q.Mul(1.0 / 64.0)
	result := new(concepts.Vector3)
	result[0] = dbg[0] - math.Floor(dbg[0])
	result[1] = dbg[1] - math.Floor(dbg[1])
	result[2] = dbg[2] - math.Floor(dbg[2])
	return result
}

func (le *LightElement) Get(wall bool) *concepts.Vector3 {
	//return le.Debug(wall)
	result := &le.Lightmap[le.MapIndex]
	if le.LightmapAge[le.MapIndex]+constants.MaxLightmapAge >= le.Config.Frame ||
		concepts.RngXorShift64()%constants.LightmapRefreshDither > 0 {
		return result
	}

	var q = new(concepts.Vector3)
	if !wall {
		le.Sector.LightmapAddressToWorld(q, le.MapIndex, le.Normal[2] > 0)
	} else {
		//log.Printf("Lightmap element doesn't exist: %v, %v, %v\n", le.Sector.Name, le.MapIndex, le.Segment.Name)
		le.Segment.LightmapAddressToWorld(q, le.MapIndex)
	}
	*result = le.Calculate(q)
	le.LightmapAge[le.MapIndex] = le.Config.Frame
	return result
}

// lightVisible determines whether a given light is visible from a world location.
func (le *LightElement) lightVisible(p *concepts.Vector3, mob *core.Mob) bool {
	// Always check the starting sector
	if le.lightVisibleFromSector(p, mob, le.Sector) {
		return true
	}

	for _, seg := range le.Sector.Segments {
		if seg.AdjacentSector.Nil() {
			continue
		}
		d2 := seg.AdjacentSegment.DistanceToPoint2(p.To2D())
		if d2 >= constants.LightGrid*constants.LightGrid {
			continue
		}

		floorZ, ceilZ := seg.AdjacentSegment.Sector.SlopedZRender(p.To2D())
		if p[2]-ceilZ > constants.LightGrid || floorZ-p[2] > constants.LightGrid {
			continue
		}
		if le.lightVisibleFromSector(p, mob, seg.AdjacentSegment.Sector) {
			return true
		}
	}

	return false
}

var debugSectorName string = "Starting" // "be4bqmfvn27mek306btg"

// TODO: Make more general
// lightVisibleFromSector determines whether a given light is visible from a world location.
func (le *LightElement) lightVisibleFromSector(p *concepts.Vector3, mob *core.Mob, sector *core.Sector) bool {
	lightPos := &mob.Pos.Render

	debugLighting := false
	if constants.DebugLighting {
		dbgName := concepts.NamedFromDb(le.DB.EntityRef(sector.Entity)).Name
		debugWallCheck := le.Normal[2] == 0
		debugLighting = constants.DebugLighting && debugWallCheck && dbgName == debugSectorName
	}

	if debugLighting {
		log.Printf("lightVisible: world=%v, light=%v\n", p.StringHuman(), lightPos)
	}
	delta := &concepts.Vector3{lightPos[0], lightPos[1], lightPos[2]}
	delta.SubSelf(p)
	maxDist2 := delta.Length2()
	// Is the point right next to the light? Visible by definition.
	if maxDist2 <= mob.BoundingRadius*mob.BoundingRadius {
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
		var next *core.Sector
		// Since our sectors can be concave, we can't just go through the first portal we find,
		// we have to go through the NEAREST one. Use dist2 to keep track...
		dist2 := maxDist2
		for _, seg := range sector.Segments {
			// Don't occlude the world location with the segment it's located on
			// Segment facing backwards from our ray? skip it.
			if (le.Normal[2] == 0 && seg == le.Segment) || delta.To2D().Dot(&seg.Normal) > 0 {
				if debugLighting {
					log.Printf("Ignoring segment [or behind] for seg %v|%v\n", seg.P.StringHuman(), seg.Next.P.StringHuman())
				}
				continue
			}

			// Find the intersection with this segment.
			isect := new(concepts.Vector3)
			ok := seg.Intersect3D(p, lightPos, isect)
			if !ok {
				if debugLighting {
					log.Printf("No intersection for seg %v|%v\n", seg.P.StringHuman(), seg.Next.P.StringHuman())
				}
				continue // No intersection, skip it!
			}

			if debugLighting {
				log.Printf("Intersection for seg %v|%v = %v\n", seg.P.StringHuman(), seg.Next.P.StringHuman(), isect.StringHuman())
			}

			if seg.AdjacentSector.Nil() {
				if debugLighting {
					log.Printf("Occluded behind wall seg %v|%v\n", seg.P.StringHuman(), seg.Next.P.StringHuman())
				}
				return false // This is a wall, that means the light is occluded for sure.
			}

			// Here, we know we have an intersected portal segment. It could still be occluding the light though, since the
			// bottom/top portions could be in the way.
			floorZ, ceilZ := sector.SlopedZRender(isect.To2D())
			floorZ2, ceilZ2 := seg.AdjacentSegment.Sector.SlopedZRender(isect.To2D())
			if debugLighting {
				log.Printf("floorZ: %v, ceilZ: %v, floorZ2: %v, ceilZ2: %v\n", floorZ, ceilZ, floorZ2, ceilZ2)
			}
			if isect[2] < floorZ2 || isect[2] > ceilZ2 || isect[2] < floorZ || isect[2] > ceilZ {
				if debugLighting {
					log.Printf("Occluded by floor/ceiling gap: %v - %v\n", seg.P.StringHuman(), seg.Next.P.StringHuman())
				}
				return false // Same as wall, we're occluded.
			}

			// Get the square of the distance to the intersection (from the target point)
			idist2 := isect.Dist2(p)

			// If the difference between the intersected distance and the light distance is
			// within the bounding radius of our light, our light is right on a portal boundary and visible.
			if math.Abs(idist2-maxDist2) < mob.BoundingRadius {
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
			dbg := fmt.Sprintf("lightVisible traversed max sectors (p: %v, light: %v)", p, mob.Entity)
			le.DebugNotices.Push(dbg)
			return false
		}
		if next == nil && mob.SectorEntityRef.Nil() && sector.Entity != mob.SectorEntityRef.Entity {
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

func (le *LightElement) Calculate(world *concepts.Vector3) concepts.Vector3 {
	diffuseSum := concepts.Vector3{}

	for _, er := range le.Sector.PVL {
		light := core.LightFromDb(er)
		if light == nil || !light.Active {
			continue
		}

		mob := core.MobFromDb(er)
		diffuseLight := 1.0
		if mob != nil && mob.Active {
			delta := &mob.Pos.Render
			delta = &concepts.Vector3{delta[0], delta[1], delta[2]}
			delta.SubSelf(world)
			dist := delta.Length()
			attenuation := 1.0
			if dist != 0 {
				// Normalize
				delta.MulSelf(1.0 / dist)
				// Calculate light strength.
				if light.Attenuation > 0.0 {
					//log.Printf("%v\n", dist)
					attenuation = light.Strength / math.Pow(dist/mob.BoundingRadius+1.0, light.Attenuation)
					//attenuation = 100.0 / dist
				}
				// If it's too far away/dark, ignore it.
				if attenuation < constants.LightAttenuationEpsilon {
					//log.Printf("Too far: %v\n", world.StringHuman())
					continue
				}
				if !le.lightVisible(world, mob) {
					//log.Printf("Shadowed: %v\n", world.StringHuman())
					continue
				}
			}
			diffuseLight = le.Normal.Dot(delta) * attenuation
		}

		if diffuseLight > 0 {
			diffuseSum.AddSelf(light.Diffuse.Mul(diffuseLight))
		}
	}
	return diffuseSum
}
