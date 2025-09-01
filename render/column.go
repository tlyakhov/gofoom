// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package render

import (
	"math"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"
)

type segmentIntersection struct {
	Segment         *core.Segment
	SectorSegment   *core.SectorSegment
	RaySegIntersect concepts.Vector3
	Distance        float64
	// Horizontal texture coordinate on segment
	U float64
	// Height of floor/ceiling at current segment intersection
	IntersectionTop, IntersectionBottom float64
}

type column struct {
	// parent
	Block *block
	// Samples shaders & images
	MaterialSampler
	// Stores current segment intersection
	segmentIntersection
	// Stores light & shadow data
	LightSampler LightSampler

	// Following data is for casting rays and intersecting them
	Sector             *core.Sector
	Ray                Ray
	RaySegTest         concepts.Vector2
	RayPlane           concepts.Vector3
	LastPortalDistance float64
	LastPortalSegment  *core.SectorSegment
	// How many portals have we traversed so far?
	Depth int
	// Height of camera above ground
	CameraZ float64
	// Fake look up/down
	ShearZ float64
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
	LightLastHash       uint64
	LightLastColHashes  []uint64
	LightLastColResults []concepts.Vector3
}

func (c *column) ProjectZ(z float64) float64 {
	return z * c.ViewFix[c.ScreenX] / c.Distance
}

func (c *column) CalcScreen() {
	// Screen slice precalculation
	c.ProjectedTop = c.ProjectZ(c.IntersectionTop - c.CameraZ)
	c.ProjectedBottom = c.ProjectZ(c.IntersectionBottom - c.CameraZ)

	if c.SectorSegment != nil && c.SectorSegment.WallUVIgnoreSlope {
		c.ProjectedSectorTop = c.ProjectZ(c.Sector.Top.Z.Render - c.CameraZ)
		c.ProjectedSectorBottom = c.ProjectZ(c.Sector.Bottom.Z.Render - c.CameraZ)
	}

	screenTop := c.ScreenHeight/2 - int(math.Floor(c.ProjectedTop)) + int(c.ShearZ)
	screenBottom := c.ScreenHeight/2 - int(math.Floor(c.ProjectedBottom)) + int(c.ShearZ)
	c.ClippedTop = concepts.Clamp(screenTop, c.EdgeTop, c.EdgeBottom)
	c.ClippedBottom = concepts.Clamp(screenBottom, c.EdgeTop, c.EdgeBottom)
}

func (c *column) SampleLight(result *concepts.Vector4, lit *materials.Lit, world *concepts.Vector3, dist float64) *concepts.Vector4 {
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
		c.LightUnfiltered(world)
		return lit.Apply(result, &c.Light)
	}

	m0 := c.WorldToLightmapHash(c.Sector, world, &c.LightSampler.Normal)
	c.LightSampler.Hash = m0
	c.LightmapHashToWorld(c.Sector, &c.LightVoxelA, m0)
	// These deltas represent 0.0 - 1.0 distances within the light voxel
	dx := (world[0] - c.LightVoxelA[0]) / c.LightGrid
	dy := (world[1] - c.LightVoxelA[1]) / c.LightGrid
	dz := (world[2] - c.LightVoxelA[2]) / c.LightGrid

	/*if dx < 0 || dy < 0 || dz < 0 || dx > 1 || dy > 1 || dz > 1 {
		fmt.Printf("Lightmap filter: dx/dy/dz < 0: %v,%v,%v\n", dx, dy, dz)
		// This duplicated code is for debugging
		m0 := c.WorldToLightmapHash(c.Sector, world, &c.LightSampler.Normal)
		c.LightSampler.Hash = m0
		c.LightmapHashToWorld(c.Sector, &c.LightVoxelA, m0)
	}*/

	// We XOR with the sector entity to avoid problems across sector boundaries
	cacheHash := m0 ^ concepts.RngXorShift64(uint64(c.Sector.Entity))
	//debugVoxel := false
	if cacheHash != c.LightLastHash {
		if cacheHash == c.LightLastColHashes[c.ScreenY] && c.LightLastColHashes[c.ScreenY] != 0 {
			copy(c.LightResult[:], c.LightLastColResults[c.ScreenY*8:c.ScreenY*8+8])
		} else {
			//debugVoxel = true
			c.LightSampler.Get()
			c.LightResult[0] = c.LightSampler.Output
			c.LightLastColHashes[c.ScreenY] = cacheHash
			for i := 1; i < 8; i++ {
				/*c.LightVoxelA[2] += float64((i & 1)) * c.LightGrid
				c.LightVoxelA[1] += float64((i&2)>>1) * c.LightGrid
				c.LightVoxelA[0] += float64((i&4)>>2) * c.LightGrid
				c.LightSampler.MapIndex = c.WorldToLightmapAddress(c.Sector, &c.LightVoxelA, &c.LightSampler.Normal)
				c.LightVoxelA[2] -= float64((i & 1)) * c.LightGrid
				c.LightVoxelA[1] -= float64((i&2)>>1) * c.LightGrid
				c.LightVoxelA[0] -= float64((i&4)>>2) * c.LightGrid*/
				// Some bit shifting to generate our light voxel hash without
				// branches. See LightmapHashToWorld for details
				c.LightSampler.Hash = m0 + uint64(i&1)<<16 + uint64(i&2)<<(32-1) + uint64(i&4)<<(48-2)
				c.LightSampler.Get()
				c.LightResult[i] = c.LightSampler.Output
			}
			copy(c.LightLastColResults[c.ScreenY*8:c.ScreenY*8+8], c.LightResult[:])
		}
		c.LightLastHash = cacheHash
	}

	// Bilinear interpolation of R,G,B components
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

func (c *column) LightUnfiltered(world *concepts.Vector3) {
	/*
		Fun dithered look, maybe leverage as an effect later?
		jitter := *world
		jitter[0] += (rand.Float64() - 0.5) * constants.LightGrid
		jitter[1] += (rand.Float64() - 0.5) * constants.LightGrid
		jitter[2] += (rand.Float64() - 0.5) * constants.LightGrid
	*/
	c.LightSampler.Hash = c.WorldToLightmapHash(c.Sector, world, &c.LightSampler.Normal)

	// We XOR with the sector entity to avoid problems across sector boundaries
	cacheHash := c.LightSampler.Hash ^ concepts.RngXorShift64(uint64(c.Sector.Entity))
	if cacheHash != c.LightLastHash {
		if cacheHash == c.LightLastColHashes[c.ScreenY] && c.LightLastColHashes[c.ScreenY] != 0 {
			copy(c.LightResult[:], c.LightLastColResults[c.ScreenY*8:c.ScreenY*8+8])
		} else {
			c.LightSampler.Get()
			c.LightResult[0] = c.LightSampler.Output
			c.LightLastColHashes[c.ScreenY] = cacheHash
			for i := 1; i < 8; i++ {
				// Some bit shifting to generate our light voxel hash without
				// branches. See LightmapHashToWorld for details
				c.LightSampler.Hash = c.LightSampler.Hash + uint64(i&1)<<16 + uint64(i&2)<<(32-1) + uint64(i&4)<<(48-2)
				c.LightResult[i] = c.LightSampler.Output
			}
			copy(c.LightLastColResults[c.ScreenY*8:c.ScreenY*8+8], c.LightResult[:])
		}
		c.LightLastHash = cacheHash
	}

	c.Light[0] = c.LightResult[0][0]
	c.Light[1] = c.LightResult[0][1]
	c.Light[2] = c.LightResult[0][2]
	c.Light[3] = 1
}
