// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package render

import (
	"fmt"

	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/selection"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/render/state"
)

func ceilingPick(s *state.Column) {
	if s.ScreenY >= s.EdgeTop && s.ScreenY < s.ClippedTop {
		s.PickedSelection = append(s.PickedSelection, selection.SelectableFromCeil(s.Sector))
	}
}

func floorPick(s *state.Column) {
	if s.ScreenY >= s.ClippedBottom && s.ScreenY < s.EdgeBottom {
		s.PickedSelection = append(s.PickedSelection, selection.SelectableFromFloor(s.Sector))
	}
}

// planes renders the top/bottom (ceiling/floor) portion of a slice.
func planes(c *state.Column, plane *core.SectorPlane) {
	mat := plane.Surface.Material
	extras := plane.Surface.ExtraStages
	c.MaterialSampler.Initialize(mat, extras)
	transform := plane.Surface.Transform.Render
	sectorMin := &c.Sector.Min
	sectorMax := &c.Sector.Max

	sw := (sectorMax[0] - sectorMin[0])
	sh := (sectorMax[1] - sectorMin[1])
	sw, sh = (transform[0]*sw+transform[2]*sh+transform[4])*c.ViewFix[c.ScreenX],
		(transform[1]*sw+transform[3]*sh+transform[5])*c.ViewFix[c.ScreenX]

	// Because of our sloped ceilings, we can't use simple linear interpolation
	// to calculate the distance or world position of the ceiling sample, we
	// have to do a ray-plane intersection.	Thankfully, the only expensive
	// operation is a square root to get the distance.
	// We could have a fast path for non-sloped cases, but it only saves a few
	// math ops and adds branches, so seems unnecessary.
	world := concepts.Vector3{}
	planeRayDelta := concepts.Vector3{
		c.Sector.Segments[0].P[0] - c.Ray.Start[0],
		c.Sector.Segments[0].P[1] - c.Ray.Start[1],
		*plane.Z.Render - c.CameraZ}
	// Top (ceiling)
	start := c.EdgeTop
	end := c.ClippedTop
	if plane == &c.Sector.Bottom {
		start = c.ClippedBottom
		end = c.EdgeBottom
	}
	for c.ScreenY = start; c.ScreenY < end; c.ScreenY++ {
		c.RayPlane[2] = float64(c.ScreenHeight/2 - c.ScreenY)
		screenIndex := uint32(c.ScreenX + c.ScreenY*c.ScreenWidth)
		denom := plane.Normal.Dot(&c.RayPlane)
		if denom == 0 {
			c.FrameBuffer[screenIndex].AddPreMulColorSelf(&concepts.Vector4{1, 0, 0, 1})
			continue
		}

		t := planeRayDelta.Dot(&plane.Normal) / denom
		if t <= 0 {
			c.FrameBuffer[screenIndex].AddPreMulColorSelf(&concepts.Vector4{1, 1, 0, 1})
			dbg := fmt.Sprintf("%v plane t <= 0", c.Sector.Entity)
			c.Player.Notices.Push(dbg)
			continue
		}

		world[2] = c.RayPlane[2] * t
		world[1] = c.RayPlane[1] * t
		world[0] = c.RayPlane[0] * t
		distToPlane := world.Length()
		dist2 := world.To2D().Length2()

		if distToPlane > c.ZBuffer[screenIndex] || dist2 > c.Distance*c.Distance {
			c.FrameBuffer[screenIndex].AddPreMulColorSelf(&concepts.Vector4{1, 0, 0, 1})
			continue
		}

		if distToPlane <= 0 {
			c.FrameBuffer[screenIndex].AddPreMulColorSelf(&concepts.Vector4{1, 0, 0, 1})
			continue
		}

		world[0] += c.Ray.Start[0]
		world[1] += c.Ray.Start[1]
		world[2] += c.CameraZ

		tx := (world[0] - sectorMin[0]) / (sectorMax[0] - sectorMin[0])
		ty := (world[1] - sectorMin[1]) / (sectorMax[1] - sectorMin[1])

		if mat != 0 {
			c.MaterialSampler.NU = tx
			c.MaterialSampler.NV = ty
			c.MaterialSampler.U = transform[0]*c.MaterialSampler.NU + transform[2]*c.MaterialSampler.NV + transform[4]
			c.MaterialSampler.V = transform[1]*c.MaterialSampler.NU + transform[3]*c.MaterialSampler.NV + transform[5]
			c.ScaleW = uint32(sw / distToPlane)
			c.ScaleH = uint32(sh / distToPlane)
			c.SampleMaterial(extras)
			c.SampleLight(&c.MaterialSampler.Output, mat, &world, distToPlane)
		}
		c.FrameBuffer[screenIndex].AddPreMulColorSelf(&c.MaterialSampler.Output)
		c.ZBuffer[screenIndex] = distToPlane
	}
}
