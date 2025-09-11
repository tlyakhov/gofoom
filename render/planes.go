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

func planePick(block *block, plane *core.SectorPlane) {
	start := block.EdgeTop
	end := block.ClippedTop
	s := selection.SelectableFromCeil(block.Sector)
	if plane == &block.Sector.Bottom {
		start = block.ClippedBottom
		end = block.EdgeBottom
		s = selection.SelectableFromFloor(block.Sector)
	}

	if block.ScreenY < start || block.ScreenY >= end {
		return
	}

	planeRayDelta := concepts.Vector3{
		block.Sector.Segments[0].P[0] - block.Ray.Start[0],
		block.Sector.Segments[0].P[1] - block.Ray.Start[1],
		plane.Z.Render - block.CameraZ}

	block.RayPlane[2] = float64(block.ScreenHeight/2 - block.ScreenY + int(block.ShearZ))
	denom := plane.Normal.Dot(&block.RayPlane)
	if denom == 0 {
		return
	}

	t := planeRayDelta.Dot(&plane.Normal) / denom
	if t <= 0 {
		return
	}

	block.PickResult.World[0] = block.Ray.Start[0] + block.RayPlane[0]*t
	block.PickResult.World[1] = block.Ray.Start[1] + block.RayPlane[1]*t
	block.PickResult.World[2] = block.CameraZ + block.RayPlane[2]*t
	block.PickResult.Normal = plane.Normal
	block.PickResult.Selection = append(block.PickResult.Selection, s)

	if plane.Normal[2] == -1 || plane.Normal[2] == 1 {
		// This avoid precision errors for flat floors
		block.PickResult.World[2] = plane.Z.Render
	}
}

// planes renders the top/bottom (ceiling/floor) portion of a slice.
func planes(block *block, plane *core.SectorPlane) {
	mat := plane.Surface.Material
	lit := materials.GetLit(plane.Surface.Material)
	extras := plane.Surface.ExtraStages
	block.MaterialSampler.Initialize(mat, extras)
	transform := plane.Surface.Transform.Render

	sectorWidth := (block.Sector.Max[0] - block.Sector.Min[0])
	sectorDepth := (block.Sector.Max[1] - block.Sector.Min[1])
	screenSpaceSectorWidth, screenSpaceSectorDepth :=
		(transform[0]*sectorWidth+transform[2]*sectorDepth+transform[4])*block.ViewFix[block.ScreenX],
		(transform[1]*sectorWidth+transform[3]*sectorDepth+transform[5])*block.ViewFix[block.ScreenX]

	// Because of our sloped ceilings, we can't use simple linear interpolation
	// to calculate the distance or world position of the ceiling sample, we
	// have to do a ray-plane intersection.	Thankfully, the only expensive
	// operation is a square root to get the distance.
	// We could have a fast path for non-sloped cases, but it only saves a few
	// math ops and adds branches, so seems unnecessary.
	world := concepts.Vector3{}
	planeRayDelta := concepts.Vector3{
		block.Sector.Segments[0].P[0] - block.Ray.Start[0],
		block.Sector.Segments[0].P[1] - block.Ray.Start[1],
		plane.Z.Render - block.CameraZ}
	// Top (ceiling)
	start := block.EdgeTop
	end := block.ClippedTop
	if plane == &block.Sector.Bottom {
		start = block.ClippedBottom
		end = block.EdgeBottom
	}
	for block.ScreenY = start; block.ScreenY < end; block.ScreenY++ {
		block.RayPlane[2] = float64(block.ScreenHeight/2 - block.ScreenY + int(block.ShearZ))
		screenIndex := uint32(block.ScreenX + block.ScreenY*block.ScreenWidth)
		denom := plane.Normal.Dot(&block.RayPlane)
		if denom == 0 {
			concepts.BlendColors(&block.FrameBuffer[screenIndex], &concepts.Vector4{1, 0, 0, 1}, 1)
			continue
		}

		t := planeRayDelta.Dot(&plane.Normal) / denom
		if t <= 0 {
			concepts.BlendColors(&block.FrameBuffer[screenIndex], &concepts.Vector4{1, 1, 0, 1}, 1)
			dbg := fmt.Sprintf("%v plane t <= 0", block.Sector.Entity)
			block.Player.Notices.Push(dbg)
			continue
		}

		world[2] = block.RayPlane[2] * t
		world[1] = block.RayPlane[1] * t
		world[0] = block.RayPlane[0] * t
		distToPlane := world.Length()
		distSq := world.To2D().Length2()

		if distToPlane > block.ZBuffer[screenIndex] || distSq > block.Distance*block.Distance+constants.IntersectEpsilon {
			concepts.BlendColors(&block.FrameBuffer[screenIndex], &concepts.Vector4{1, 0, 1, 1}, 1)
			continue
		}

		if distToPlane <= 0 {
			concepts.BlendColors(&block.FrameBuffer[screenIndex], &concepts.Vector4{0, 1, 0, 1}, 1)
			continue
		}

		world[0] += block.Ray.Start[0]
		world[1] += block.Ray.Start[1]
		switch plane.Normal[2] {
		case 1, -1:
			// This avoid precision errors for flat floors
			world[2] = plane.Z.Render
		default:
			world[2] += block.CameraZ
			//world[2] = plane.ZAt(dynamic.DynamicRender, world.To2D())
		}

		if mat != 0 {
			block.MaterialSampler.NU = (world[0] - block.Sector.Min[0]) / sectorWidth
			block.MaterialSampler.NV = (world[1] - block.Sector.Min[1]) / sectorDepth
			block.MaterialSampler.U = transform[0]*block.MaterialSampler.NU + transform[2]*block.MaterialSampler.NV + transform[4]
			block.MaterialSampler.V = transform[1]*block.MaterialSampler.NU + transform[3]*block.MaterialSampler.NV + transform[5]
			block.ScaleW = uint32(screenSpaceSectorWidth / distToPlane)
			block.ScaleH = uint32(screenSpaceSectorDepth / distToPlane)
			block.SampleMaterial(extras)
			if lit != nil {
				block.SampleLight(&block.MaterialSampler.Output, lit, &world, distToPlane)
			}
		}
		concepts.BlendColors(&block.FrameBuffer[screenIndex], &block.MaterialSampler.Output, 1)
		block.ZBuffer[screenIndex] = distToPlane
	}
}
