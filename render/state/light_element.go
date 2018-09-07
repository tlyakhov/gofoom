package state

import (
	"fmt"
	"math"

	"github.com/tlyakhov/gofoom/behaviors"
	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/constants"
	"github.com/tlyakhov/gofoom/core"
)

type LightElement struct {
	Normal   concepts.Vector3
	Sector   *core.PhysicalSector
	Segment  *core.Segment
	MapIndex uint32
	Lightmap []concepts.Vector3
}

func (le *LightElement) Get(wall bool) concepts.Vector3 {
	if le.MapIndex < 0 {
		le.MapIndex = 0
	}

	v := le.Lightmap[le.MapIndex]
	if v.X < 0 {
		var q concepts.Vector3
		if !wall {
			q = le.Sector.LightmapAddressToWorld(le.MapIndex, le.Normal.Z > 0)
		} else {
			q = le.Segment.LightmapAddressToWorld(le.MapIndex)
		}
		le.Lightmap[le.MapIndex] = le.Calculate(q)
	}
	return le.Lightmap[le.MapIndex]
}

// lightVisible determines whether a given light is visible from a world location.
func (le *LightElement) lightVisible(p concepts.Vector3, e *core.PhysicalEntity) bool {
	p2d := p.To2D()
	sector := le.Sector
	// Our point could be outside of the current sector...
	if !sector.IsPointInside2D(p2d) {
		for _, seg := range sector.Segments {
			/*if seg.WhichSide(p2d) > 0 {
				continue
			}*/
			if seg.AdjacentSector != nil && seg.AdjacentSector.IsPointInside2D(p2d) {
				floorZ, ceilZ := seg.AdjacentSector.Physical().CalcFloorCeilingZ(p2d)
				if ceilZ-floorZ <= constants.IntersectEpsilon {
					return false
				}
				sector = seg.AdjacentSector.Physical()
				break
			}
			closest := seg.ClosestToPoint(p2d)
			delta := closest.Sub(p2d)
			d := delta.Length()
			if d > constants.LightGrid*2 {
				continue
			}

			if seg.WhichSide(p2d) > 0 {
				p2d := p2d.Add(delta.Norm().Mul(-d + constants.LightGrid))
				p.X = p2d.X
				p.Y = p2d.Y
			} else {
				p2d := p2d.Add(delta.Norm().Mul(-d - constants.LightGrid))
				p.X = p2d.X
				p.Y = p2d.Y
			}
		}
	}

	floorZ, ceilZ := sector.CalcFloorCeilingZ(p2d)
	if ceilZ-floorZ <= constants.IntersectEpsilon {
		return false
	}

	delta := e.Pos.Sub(p)
	maxDist := delta.Length2()
	// Is the point right next to the light? Visible by definition.
	if maxDist <= e.BoundingRadius*e.BoundingRadius {
		return true
	}

	// The outer loop traverses portals starting from the sector our target point is in,
	// and finishes in the sector our light is in (unless occluded)
	depth := 0 // We keep track of portaling depth to avoid infinite traversal in weird cases.
	prevDist := -1.0
	for sector != nil {
		var next *core.PhysicalSector
		// Since our sectors can be concave, we can't just go through the first portal we find,
		// we have to go through the NEAREST one. Use dist to keep track...
		dist := maxDist
		for _, seg := range sector.Segments {
			// Segment facing backwards from our ray? skip it.
			if delta.To2D().Dot(seg.Normal) > 0 {
				continue
			}
			// Find the intersection with this segment.
			isect, ok := seg.Intersect3D(p, e.Pos)
			if !ok {
				continue // No intersection, skip it!
			}
			// Get the square of the distance to the intersection (from the target point)
			d := isect.Dist2(p)

			if seg.AdjacentSector == nil {
				return false // This is a wall, that means the light is occluded for sure.
			} else if d > dist || d <= prevDist {
				// If the current intersection point is farther than one we already have for this sector,
				// or BEHIND the last one, we have a concavity. Keep looking.
				continue
			}
			// Here, we know we have an intersected portal segment. It could still be occluding the light though, since the
			// bottom/top portions could be in the way.
			floorZ, ceilZ := sector.CalcFloorCeilingZ(isect.To2D())
			if isect.Z < floorZ || isect.Z > ceilZ {
				return false // Same as wall, we're occluded.
			}
			// We're in the clear! Move to the next adjacent sector.
			dist = d
			next = seg.AdjacentSector.Physical()
		}
		prevDist = dist
		depth++
		if depth > 100 { // Avoid infinite looping.
			fmt.Printf("warning: lightVisible traversed > 100 sectors.\n")
			return false
		}
		sector = next
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

			// Normalize
			delta = delta.Mul(1.0 / dist)
			// Calculate light strength.
			attenuation := 1.0
			if lb.Attenuation > 0.0 {
				//fmt.Printf("%v\n", dist)
				attenuation = lb.Strength / math.Pow(dist/lightEntity.Physical().BoundingRadius+1.0, lb.Attenuation)
				//attenuation = 100.0 / dist
			}
			// If it's too far away/dark, ignore it.
			if attenuation < constants.LightAttenuationEpsilon || !le.lightVisible(world, lightEntity.Physical()) {
				continue
			}
			diffuseLight := le.Normal.Dot(delta) * attenuation
			if diffuseLight > 0 {
				diffuseSum = diffuseSum.Add(lb.Diffuse.Mul(diffuseLight))
			}
		}
	}
	return diffuseSum
}
