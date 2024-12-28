// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package render

import (
	"fmt"

	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/components/selection"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
)

func ceilingPick(s *column) {
	if s.ScreenY >= s.EdgeTop && s.ScreenY < s.ClippedTop {
		s.PickedSelection = append(s.PickedSelection, selection.SelectableFromCeil(s.Sector))
	}
}

func floorPick(s *column) {
	if s.ScreenY >= s.ClippedBottom && s.ScreenY < s.EdgeBottom {
		s.PickedSelection = append(s.PickedSelection, selection.SelectableFromFloor(s.Sector))
	}
}

// planes renders the top/bottom (ceiling/floor) portion of a slice.
func planes(c *column, plane *core.SectorPlane) {
	mat := plane.Surface.Material
	lit := materials.GetLit(c.ECS, plane.Surface.Material)
	extras := plane.Surface.ExtraStages
	c.MaterialSampler.Initialize(mat, extras)
	transform := plane.Surface.Transform.Render

	sectorWidth := (c.Sector.Max[0] - c.Sector.Min[0])
	sectorDepth := (c.Sector.Max[1] - c.Sector.Min[1])
	screenSpaceSectorWidth, screenSpaceSectorDepth :=
		(transform[0]*sectorWidth+transform[2]*sectorDepth+transform[4])*c.ViewFix[c.ScreenX],
		(transform[1]*sectorWidth+transform[3]*sectorDepth+transform[5])*c.ViewFix[c.ScreenX]

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
			concepts.BlendColors(&c.FrameBuffer[screenIndex], &concepts.Vector4{1, 0, 0, 1}, 1)
			continue
		}

		t := planeRayDelta.Dot(&plane.Normal) / denom
		if t <= 0 {
			concepts.BlendColors(&c.FrameBuffer[screenIndex], &concepts.Vector4{1, 1, 0, 1}, 1)
			dbg := fmt.Sprintf("%v plane t <= 0", c.Sector.Entity)
			c.Player.Notices.Push(dbg)
			continue
		}

		world[2] = c.RayPlane[2] * t
		world[1] = c.RayPlane[1] * t
		world[0] = c.RayPlane[0] * t
		distToPlane := world.Length()
		dist2 := world.To2D().Length2()

		if distToPlane > c.ZBuffer[screenIndex] || dist2 > c.Distance*c.Distance+constants.IntersectEpsilon {
			concepts.BlendColors(&c.FrameBuffer[screenIndex], &concepts.Vector4{1, 0, 1, 1}, 1)
			continue
		}

		if distToPlane <= 0 {
			concepts.BlendColors(&c.FrameBuffer[screenIndex], &concepts.Vector4{0, 1, 0, 1}, 1)
			continue
		}

		world[0] += c.Ray.Start[0]
		world[1] += c.Ray.Start[1]
		switch plane.Normal[2] {
		case 1, -1:
			world[2] = *plane.Z.Render
		default:
			world[2] += c.CameraZ
			//world[2] = plane.ZAt(dynamic.DynamicRender, world.To2D())
		}

		if mat != 0 {
			c.MaterialSampler.NU = (world[0] - c.Sector.Min[0]) / sectorWidth
			c.MaterialSampler.NV = (world[1] - c.Sector.Min[1]) / sectorDepth
			c.MaterialSampler.U = transform[0]*c.MaterialSampler.NU + transform[2]*c.MaterialSampler.NV + transform[4]
			c.MaterialSampler.V = transform[1]*c.MaterialSampler.NU + transform[3]*c.MaterialSampler.NV + transform[5]
			c.ScaleW = uint32(screenSpaceSectorWidth / distToPlane)
			c.ScaleH = uint32(screenSpaceSectorDepth / distToPlane)
			c.SampleMaterial(extras)
			if lit != nil {
				c.SampleLight(&c.MaterialSampler.Output, lit, &world, distToPlane)
			}
		}
		concepts.BlendColors(&c.FrameBuffer[screenIndex], &c.MaterialSampler.Output, 1)
		c.ZBuffer[screenIndex] = distToPlane
	}
}
