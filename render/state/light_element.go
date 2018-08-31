package state

import (
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
	U, V     float64
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
		le.Calculate(q)
	}
	return le.Lightmap[le.MapIndex]
}

func (le *LightElement) lightVisible(p concepts.Vector3, e *core.PhysicalEntity) bool {
	//fmt.Println(p.String())
	rayLength := p.Dist(e.Pos)
	if rayLength == 0 {
		return true
	}

	sector := le.Sector
	visited := make(map[string]bool)

	for sector != nil {
		visited[sector.ID] = true
		var next *core.PhysicalSector
		for _, seg := range sector.Segments {
			delta := e.Pos.Sub(p)
			if delta.To2D().Dot(seg.Normal) > 0 {
				continue
			}
			isect, ok := seg.Intersect3D(p, e.Pos)
			if !ok {
				continue
			} else if ok && seg.AdjacentSector == nil {
				return false
			} else if visited[seg.AdjacentSector.GetBase().ID] {
				continue
			}
			floorZ, ceilZ := seg.Sector.Physical().CalcFloorCeilingZ(isect.To2D())
			if isect.Z < floorZ || isect.Z > ceilZ {
				return false
			}
			next = seg.AdjacentSector.Physical()
		}
		if next == nil {
			break
		}
		sector = next
	}
	return true
}

func (le *LightElement) Calculate(world concepts.Vector3) {
	w2d := world.To2D()
	p := world
	if false && !le.Sector.IsPointInside2D(w2d) {
		for _, seg := range le.Sector.Segments {
			if seg.WhichSide(w2d) > 0 {
				continue
			}
			closest := seg.ClosestToPoint(w2d)
			v := closest.Sub(w2d)
			d := v.Length()
			if d > constants.LightGrid*2 {
				continue
			} else if d > constants.IntersectEpsilon {
				p = p.Add(v.Norm().Mul(d + constants.LightGrid*0.5).To3D())
			} else {
				p = p.Add(seg.Normal.Mul(constants.LightGrid * 0.5).To3D())
			}
		}
		// Debugging...
		le.Lightmap[le.MapIndex] = concepts.Vector3{1, 0, 0}
		return
	}

	diffuseSum := concepts.Vector3{}

	for _, lightEntity := range le.Sector.PVL {
		if !lightEntity.Physical().Active || !le.lightVisible(p, lightEntity.Physical()) {
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
				attenuation = lb.Strength / math.Pow(dist/lightEntity.Physical().BoundingRadius+1.0, lb.Attenuation)
			}
			// If it's too far away/dark, ignore it.
			if attenuation < constants.LightAttenuationEpsilon {
				continue
			}
			diffuseLight := le.Normal.Dot(delta) * attenuation
			if diffuseLight > 0 {
				diffuseSum = diffuseSum.Add(lb.Diffuse.Mul(diffuseLight))
			}
		}
	}
	le.Lightmap[le.MapIndex] = diffuseSum
}
