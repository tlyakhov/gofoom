// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package state

import (
	"fmt"
	"math"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/components/selection"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/containers"
	"tlyakhov/gofoom/ecs"
)

type EntityWithDist2 struct {
	Body            *core.Body
	Visible         *materials.Visible
	InternalSegment *core.InternalSegment
	Sector          *core.Sector
	Dist2           float64
}

type SegmentIntersection struct {
	Segment         *core.Segment
	SectorSegment   *core.SectorSegment
	RaySegIntersect concepts.Vector3
	Distance        float64
	// Horizontal texture coordinate on segment
	U float64
	// Height of floor/ceiling at current segment intersection
	IntersectionTop, IntersectionBottom float64
}

type Column struct {
	// Global rendering configuration
	*Config
	// Samples shaders & images
	MaterialSampler
	// Stores current segment intersection
	*SegmentIntersection
	// Stores light & shadow data
	LightSampler LightSampler
	// Pre-allocated stack of past intersections, for speed
	Visited []SegmentIntersection
	// Pre-allocated stack of nested columns for portals
	PortalColumns []Column
	// Pre-allocated slice for sorting bodies and internal segments
	Sectors containers.Set[*core.Sector]
	// Following data is for casting rays and intersecting them
	Sector             *core.Sector
	Ray                *Ray
	RaySegTest         concepts.Vector2
	RayPlane           concepts.Vector3
	LastPortalDistance float64
	// How many portals have we traversed so far?
	Depth int
	// Height of camera above ground
	CameraZ float64
	// Scaled screenspace boundaries of current column (unclipped)
	EdgeTop, EdgeBottom int
	// Projected height of floor/ceiling at current segment intersection
	ProjectedTop, ProjectedBottom float64
	// Projected height of sector floor/ceiling if wall ignores slope
	ProjectedSectorTop, ProjectedSectorBottom float64
	// Screen-space coordinates clipped to edges
	ClippedTop, ClippedBottom int
	// Lighting cache
	Light               concepts.Vector4
	LightVoxelA         concepts.Vector3
	LightResult         [8]concepts.Vector3
	LightLastIndex      uint64
	LightLastColIndices []uint64
	LightLastColResults []concepts.Vector3
	// For picking things in editor
	Pick            bool
	PickedSelection []*selection.Selectable
}

func (c *Column) ProjectZ(z float64) float64 {
	return z * c.ViewFix[c.ScreenX] / c.Distance
}

func (c *Column) CalcScreen() {
	// Screen slice precalculation
	c.ProjectedTop = c.ProjectZ(c.IntersectionTop - c.CameraZ)
	c.ProjectedBottom = c.ProjectZ(c.IntersectionBottom - c.CameraZ)

	if c.SectorSegment != nil && c.SectorSegment.WallUVIgnoreSlope {
		c.ProjectedSectorTop = c.ProjectZ(*c.Sector.Top.Z.Render - c.CameraZ)
		c.ProjectedSectorBottom = c.ProjectZ(*c.Sector.Bottom.Z.Render - c.CameraZ)
	}

	screenTop := c.ScreenHeight/2 - int(math.Floor(c.ProjectedTop))
	screenBottom := c.ScreenHeight/2 - int(math.Floor(c.ProjectedBottom))
	c.ClippedTop = concepts.Clamp(screenTop, c.EdgeTop, c.EdgeBottom)
	c.ClippedBottom = concepts.Clamp(screenBottom, c.EdgeTop, c.EdgeBottom)
}

func (c *Column) SampleLight(result *concepts.Vector4, material ecs.Entity, world *concepts.Vector3, dist float64) *concepts.Vector4 {
	lit := materials.GetLit(c.ECS, material)

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
	if dist > float64(c.ScreenWidth)*c.LightGrid*0.25 {
		c.LightUnfiltered(&c.Light, world)
		return lit.Apply(result, &c.Light)
	}

	extraHash := uint16(c.LightSampler.Type)
	if c.LightSampler.Type == LightSamplerWall {
		extraHash += c.Segment.LightExtraHash
	}

	m0 := c.WorldToLightmapAddress(c.Sector, world, extraHash)
	c.LightSampler.MapIndex = m0
	c.LightmapAddressToWorld(c.Sector, &c.LightVoxelA, m0)
	// These deltas represent 0.0 - 1.0 distances within the light voxel
	dx := (world[0] - c.LightVoxelA[0]) / c.LightGrid
	dy := (world[1] - c.LightVoxelA[1]) / c.LightGrid
	dz := (world[2] - c.LightVoxelA[2]) / c.LightGrid

	if dx < 0 || dy < 0 || dz < 0 {
		fmt.Printf("Lightmap filter: dx/dy/dz < 0: %v,%v,%v\n", dx, dy, dz)
	}

	//debugVoxel := false
	if m0 != c.LightLastIndex {
		if m0 == c.LightLastColIndices[c.ScreenY] && c.LightLastColIndices[c.ScreenY] != 0 {
			copy(c.LightResult[:], c.LightLastColResults[c.ScreenY*8:c.ScreenY*8+8])
		} else {
			//debugVoxel = true
			c.LightSampler.Get()
			c.LightResult[0] = c.LightSampler.Output
			c.LightLastColIndices[c.ScreenY] = m0
			for i := 1; i < 8; i++ {
				// Some bit shifting to generate our light voxel
				// addresses without ifs. See LightmapAddressToWorld for details
				c.LightSampler.MapIndex = m0 + uint64(i&1)<<16 + uint64(i&2)<<(32-1) + uint64(i&4)<<(48-2)
				c.LightSampler.Get()
				c.LightResult[i] = c.LightSampler.Output
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

	/*if debugVoxel {
		c.Light[0] = 1
		c.Light[1] = 0
		c.Light[2] = 0
	}*/

	return lit.Apply(result, &c.Light)
}

func (c *Column) LightUnfiltered(result *concepts.Vector4, world *concepts.Vector3) *concepts.Vector4 {
	extraHash := uint16(c.LightSampler.Type)
	if c.LightSampler.Type == LightSamplerWall {
		extraHash += c.Segment.LightExtraHash
	}
	/*
		Fun dithered look, maybe leverage as an effect later?
		jitter := *world
		jitter[0] += (rand.Float64() - 0.5) * constants.LightGrid
		jitter[1] += (rand.Float64() - 0.5) * constants.LightGrid
		jitter[2] += (rand.Float64() - 0.5) * constants.LightGrid
	*/
	c.LightSampler.MapIndex = c.WorldToLightmapAddress(c.Sector, world, extraHash)

	r00 := c.LightSampler.Get()
	result[0] = r00[0]
	result[1] = r00[1]
	result[2] = r00[2]
	result[3] = 1
	return result
}
