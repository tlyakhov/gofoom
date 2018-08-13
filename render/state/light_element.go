package state

import (
	"math"

	"github.com/tlyakhov/gofoom/behaviors"
	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/constants"
	"github.com/tlyakhov/gofoom/core"
)

type LightElement struct {
	Normal        concepts.Vector3
	Sector        *core.PhysicalSector
	Segment       *core.Segment
	U, V          float64
	MapIndex      uint32
	Lightmap      []concepts.Vector3
	VisibleLights map[string]*core.PhysicalEntity
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

func (le *LightElement) markVisibleLight(p concepts.Vector3, e *core.PhysicalEntity) {
	rayLength := p.Dist(e.Pos)
	if rayLength == 0 {
		return
	}

	sector := le.Sector

	for sector != nil {
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
				delete(le.VisibleLights, e.ID)
				continue
			} else if isect.Z < sector.BottomZ && isect.Z > sector.TopZ {
				delete(le.VisibleLights, e.ID)
				continue
			}
			next = seg.AdjacentSector.Physical()
		}
		if next == nil {
			break
		}
		sector = next
	}
}

func (le *LightElement) MarkVisibleLights(p concepts.Vector3) {
	le.VisibleLights = make(map[string]*core.PhysicalEntity)
	for _, l := range le.Sector.PVSLights {
		if !l.Physical().Active {
			continue
		}

		for _, b := range l.Behaviors() {
			_, ok := b.(*behaviors.Light)
			if !ok {
				continue
			}
			e := l.Physical()
			le.VisibleLights[e.ID] = e
			le.markVisibleLight(p, e)
		}
	}
}

func (le *LightElement) Calculate(world concepts.Vector3) {
	w2d := world.To2D()
	p := world
	if !le.Sector.IsPointInside2D(w2d) {
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
		//le.Lightmap[le.MapIndex+0] = 1
		//le.Lightmap[le.MapIndex+1] = 0
		//le.Lightmap[le.MapIndex+2] = 0
		//return
	}

	le.MarkVisibleLights(p)
	diffuseSum := concepts.ZeroVector3

	for _, lightEntity := range le.Sector.PVSLights {
		if !lightEntity.Physical().Active {
			continue
		}

		for _, b := range lightEntity.Behaviors() {
			lb, ok := b.(*behaviors.Light)
			if !ok {
				continue
			}
			if le.VisibleLights == nil {
				le.VisibleLights = make(map[string]*core.PhysicalEntity)
				le.VisibleLights[lightEntity.Physical().ID] = lightEntity.Physical()
				continue
			} else if _, vis := le.VisibleLights[lightEntity.Physical().ID]; !vis {
				le.VisibleLights[lightEntity.Physical().ID] = lightEntity.Physical()
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
