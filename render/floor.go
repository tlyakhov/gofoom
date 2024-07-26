// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package render

import (
	"fmt"

	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/render/state"
)

func FloorPick(s *state.Column) {
	if s.ScreenY >= s.ClippedBottom && s.ScreenY < s.EdgeBottom {
		s.PickedSelection = append(s.PickedSelection, core.SelectableFromFloor(s.Sector))
	}
}

// Floor renders the floor portion of a slice.
func Floor(c *state.Column) {
	mat := c.Sector.FloorSurface.Material
	extras := c.Sector.FloorSurface.ExtraStages
	transform := c.Sector.FloorSurface.Transform.Render
	sectorMin := &c.Sector.Min
	sectorMax := &c.Sector.Max

	sw := (sectorMax[0] - sectorMin[0])
	sh := (sectorMax[1] - sectorMin[1])
	sw, sh = (transform[0]*sw+transform[2]*sh+transform[4])*c.ViewFix[c.ScreenX],
		(transform[1]*sw+transform[3]*sh+transform[5])*c.ViewFix[c.ScreenX]

	// Because of our sloped floors, we can't use simple linear interpolation
	// to calculate the distance or world position of the ceiling sample, we
	// have to do a ray-plane intersection.	Thankfully, the only expensive
	// operation is a square root to get the distance.
	// We could have a fast path for non-sloped cases, but it only saves a few
	// math ops and adds branches, so seems unnecessary.
	world := concepts.Vector3{}
	planeRayDelta := concepts.Vector3{
		c.Sector.Segments[0].P[0] - c.Ray.Start[0],
		c.Sector.Segments[0].P[1] - c.Ray.Start[1],
		*c.Sector.BottomZ.Render - c.CameraZ}
	for c.ScreenY = c.ClippedBottom; c.ScreenY < c.EdgeBottom; c.ScreenY++ {
		c.RayFloorCeil[2] = float64(c.ScreenHeight/2 - c.ScreenY)
		denom := c.Sector.FloorNormal.Dot(&c.RayFloorCeil)

		if denom == 0 {
			continue
		}

		t := planeRayDelta.Dot(&c.Sector.FloorNormal) / denom
		if t <= 0 {
			//s.Write(uint32(s.X+s.Y*s.ScreenWidth), 255)
			dbg := fmt.Sprintf("%v floor t <= 0", c.Sector.Entity)
			c.DebugNotices.Push(dbg)
			continue
		}
		world[2] = c.RayFloorCeil[2] * t
		world[1] = c.RayFloorCeil[1] * t
		world[0] = c.RayFloorCeil[0] * t
		distToFloor := world.Length()
		dist2 := world.To2D().Length2()
		screenIndex := uint32(c.ScreenX + c.ScreenY*c.ScreenWidth)

		if distToFloor > c.ZBuffer[screenIndex] || dist2 > c.Distance*c.Distance {
			continue
		}

		world[0] += c.Ray.Start[0]
		world[1] += c.Ray.Start[1]
		world[2] += c.CameraZ

		tx := (world[0] - sectorMin[0]) / (sectorMax[0] - sectorMin[0])
		ty := (world[1] - sectorMin[1]) / (sectorMax[1] - sectorMin[1])

		if mat != 0 {
			tx, ty = transform[0]*tx+transform[2]*ty+transform[4], transform[1]*tx+transform[3]*ty+transform[5]
			c.SampleShader(mat, extras, tx, ty, uint32(sw/distToFloor), uint32(sh/distToFloor))
			c.SampleLight(&c.MaterialSampler.Output, mat, &world, distToFloor)
		}
		c.FrameBuffer[screenIndex].AddPreMulColorSelf(&c.MaterialSampler.Output)
		c.ZBuffer[screenIndex] = distToFloor
	}
}
