// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package state

import (
	"fmt"
	"math"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
)

type EntityWithDist2 struct {
	concepts.Entity
	Dist2     float64
	IsSegment bool
}

type Ray struct {
	Start, End, Delta concepts.Vector2
	Angle             float64
	AngleCos          float64
	AngleSin          float64
}

func (r *Ray) Set(a float64) {
	r.Angle = a
	r.AngleSin, r.AngleCos = math.Sincos(a)
	r.End = concepts.Vector2{
		r.Start[0] + constants.MaxViewDistance*r.AngleCos,
		r.Start[1] + constants.MaxViewDistance*r.AngleSin,
	}
	r.Delta.From(&r.End).SubSelf(&r.Start)
}

func (r *Ray) AnglesFromStartEnd() {
	r.Delta.From(&r.End).SubSelf(&r.Start)
	r.Angle = math.Atan2(r.Delta[1], r.Delta[0])
	r.AngleSin, r.AngleCos = math.Sincos(r.Angle)
}

type Column struct {
	// Global rendering configuration
	*Config
	// Samples shaders & images
	MaterialSampler
	// Stores light & shadow data
	LightElement
	// Pre-allocated stack of nested columns for portals
	PortalColumns []Column
	// Pre-allocated slice for sorting bodies and internal segments
	EntitiesByDistance []EntityWithDist2
	// Following data is for casting rays and intersecting them
	Sector             *core.Sector
	Segment            *core.Segment
	SectorSegment      *core.SectorSegment
	Ray                *Ray
	RaySegTest         concepts.Vector2
	RaySegIntersect    concepts.Vector3
	RayFloorCeil       concepts.Vector3
	Distance           float64
	LastPortalDistance float64
	// Horizontal texture coordinate on segment
	U float64
	// How many portals have we traversed so far?
	Depth int
	// Height of camera above ground
	CameraZ float64
	// Height of floor/ceiling at current segment intersection
	IntersectionTop, IntersectionBottom float64
	// Scaled screenspace boundaries of current column (unclipped)
	EdgeTop, EdgeBottom int
	// Projected height of floor/ceiling at current segment intersection
	ProjectedTop, ProjectedBottom float64
	ScreenTop, ScreenBottom       int
	ClippedTop, ClippedBottom     int
	// Lightning cache
	Light               concepts.Vector4
	LightVoxelA         concepts.Vector3
	LightResult         [8]concepts.Vector3
	LightLastIndex      uint64
	LightLastColIndices []uint64
	LightLastColResults []concepts.Vector3
	// For picking things in editor
	Pick            bool
	PickedSelection []*core.Selectable
}

// This function has side effects: it fills in various fields on the Column if
// there was an intersection, and affects Column.RaySegTest even if not.
func (c *Column) IntersectSegment(segment *core.Segment, checkDist bool, twoSided bool) bool {
	// Wall is facing away from us
	if !twoSided && c.Ray.Delta.Dot(&segment.Normal) > 0 {
		return false
	}

	// Ray intersects?
	if ok := segment.Intersect2D(&c.Ray.Start, &c.Ray.End, &c.RaySegTest); !ok {
		/*	if c.Sector.Entity == 82 {
			dbg := fmt.Sprintf("No intersection %v <-> %v", segment.A.StringHuman(), segment.B.StringHuman())
			c.DebugNotices.Push(dbg)
		}*/
		return false
	}

	var dist float64
	dx := math.Abs(c.RaySegTest[0] - c.Ray.Start[0])
	dy := math.Abs(c.RaySegTest[1] - c.Ray.Start[1])
	if dy > dx {
		dist = math.Abs(dy / c.Ray.AngleSin)
	} else {
		dist = math.Abs(dx / c.Ray.AngleCos)
	}

	if checkDist && (dist > c.Distance || dist < c.LastPortalDistance) {
		return false
	}

	c.Segment = segment
	c.Distance = dist
	c.RaySegIntersect[0] = c.RaySegTest[0]
	c.RaySegIntersect[1] = c.RaySegTest[1]
	c.U = c.RaySegTest.Dist(segment.A) / segment.Length
	return true
}

func (c *Column) ProjectZ(z float64) float64 {
	return z * c.ViewFix[c.ScreenX] / c.Distance
}

func (c *Column) CalcScreen() {
	// Screen slice precalculation
	c.ProjectedTop = c.ProjectZ(c.IntersectionTop - c.CameraZ)
	c.ProjectedBottom = c.ProjectZ(c.IntersectionBottom - c.CameraZ)

	c.ScreenTop = c.ScreenHeight/2 - int(math.Floor(c.ProjectedTop))
	c.ScreenBottom = c.ScreenHeight/2 - int(math.Floor(c.ProjectedBottom))
	c.ClippedTop = concepts.Clamp(c.ScreenTop, c.EdgeTop, c.EdgeBottom)
	c.ClippedBottom = concepts.Clamp(c.ScreenBottom, c.EdgeTop, c.EdgeBottom)
}

func (c *Column) SampleLight(result *concepts.Vector4, material concepts.Entity, world *concepts.Vector3, dist float64) *concepts.Vector4 {
	lit := materials.LitFromDb(c.DB, material)

	if lit == nil {
		return result
	}

	// testing...
	/*dbg := world.Mul(1.0 / 64.0)
	result[0] = dbg[0] - math.Floor(dbg[0])
	result[1] = dbg[1] - math.Floor(dbg[1])
	result[2] = dbg[2] - math.Floor(dbg[2])
	return result*/
	// Test depth
	/*result[0] = concepts.Clamp(dist/500.0, 0.0, 1.0)
	result[1] = result[0]
	result[2] = result[0]
	return result*/

	// Don't filter far away lightmaps. Tolerate a ~2px snap-in
	if dist > float64(c.ScreenWidth)*constants.LightGrid*0.25 {
		c.LightUnfiltered(&c.Light, world)
		return lit.Apply(result, &c.Light)
	}

	flags := uint16(c.LightElement.Type)
	if c.LightElement.Type == LightElementWall {
		flags += uint16(c.LightElement.Segment.Index)
	}

	m0 := c.Sector.WorldToLightmapAddress(world, flags)
	c.LightElement.MapIndex = m0
	c.Sector.LightmapAddressToWorld(&c.LightVoxelA, m0)
	// These deltas represent 0.0 - 1.0 distances within the light voxel
	dx := (world[0] - c.LightVoxelA[0]) / constants.LightGrid
	dy := (world[1] - c.LightVoxelA[1]) / constants.LightGrid
	dz := (world[2] - c.LightVoxelA[2]) / constants.LightGrid

	if dx < 0 || dy < 0 || dz < 0 {
		fmt.Printf("Lightmap filter: dx/dy/dz < 0: %v,%v,%v\n", dx, dy, dz)
	}

	//debugVoxel := false
	if m0 != c.LightLastIndex {
		if m0 == c.LightLastColIndices[c.ScreenY] && c.LightLastColIndices[c.ScreenY] != 0 {
			copy(c.LightResult[:], c.LightLastColResults[c.ScreenY*8:c.ScreenY*8+8])
		} else {
			//debugVoxel = true
			c.LightElement.Get()
			c.LightResult[0] = c.LightElement.Output
			c.LightLastColIndices[c.ScreenY] = m0
			for i := 1; i < 8; i++ {
				// Some bit shifting to generate our light voxel
				// addresses without ifs. See LightmapAddressToWorld for details
				c.LightElement.MapIndex = m0 + uint64(i&1)<<16 + uint64(i&2)<<(32-1) + uint64(i&4)<<(48-2)
				c.LightElement.Get()
				c.LightResult[i] = c.LightElement.Output
			}
			copy(c.LightLastColResults[c.ScreenY*8:c.ScreenY*8+8], c.LightResult[:])
		}
		c.LightLastIndex = m0
	}

	for i := range 3 {
		c00 := c.LightResult[0][i]*(1.0-dx) + c.LightResult[4][i]*dx
		c01 := c.LightResult[1][i]*(1.0-dx) + c.LightResult[5][i]*dx
		c10 := c.LightResult[2][i]*(1.0-dx) + c.LightResult[6][i]*dx
		c11 := c.LightResult[3][i]*(1.0-dx) + c.LightResult[7][i]*dx
		c0 := c00*(1.0-dy) + c10*dy
		c1 := c01*(1.0-dy) + c11*dy
		c.Light[i] = c0*(1.0-dz) + c1*dz
	}
	c.Light[3] = 1

	/*	if debugVoxel {
		c.Light[0] = 1
		c.Light[1] = 0
		c.Light[2] = 0
	}*/

	return lit.Apply(result, &c.Light)
}

func (c *Column) LightUnfiltered(result *concepts.Vector4, world *concepts.Vector3) *concepts.Vector4 {
	flags := uint16(c.LightElement.Type)
	if c.LightElement.Type == LightElementWall {
		flags += uint16(c.LightElement.Segment.Index)
	}
	/*
		Fun dithered look, maybe leverage as an effect later?
		jitter := *world
		jitter[0] += (rand.Float64() - 0.5) * constants.LightGrid
		jitter[1] += (rand.Float64() - 0.5) * constants.LightGrid
		jitter[2] += (rand.Float64() - 0.5) * constants.LightGrid
	*/
	c.LightElement.MapIndex = c.Sector.WorldToLightmapAddress(world, flags)

	r00 := c.LightElement.Get()
	result[0] = r00[0]
	result[1] = r00[1]
	result[2] = r00[2]
	result[3] = 1
	return result
}

func (c *Column) ApplySample(sample *concepts.Vector4, screenIndex int, z float64) {
	sample.ClampSelf(0, 1)
	if sample[3] == 0 {
		return
	}
	if sample[3] == 1 {
		c.FrameBuffer[screenIndex] = *sample
		c.ZBuffer[screenIndex] = z
		return
	}
	dst := &c.FrameBuffer[screenIndex]
	dst[0] = dst[0]*(1.0-sample[3]) + sample[0]
	dst[1] = dst[1]*(1.0-sample[3]) + sample[1]
	dst[2] = dst[2]*(1.0-sample[3]) + sample[2]
	if sample[3] > 0.8 {
		c.ZBuffer[screenIndex] = z
	}
}
