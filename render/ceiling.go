// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package render

import (
	"fmt"

	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/render/state"
)

func ceilingPick(s *state.Column) {
	if s.ScreenY >= s.EdgeTop && s.ScreenY < s.ClippedTop {
		s.PickedSelection = append(s.PickedSelection, core.SelectableFromCeil(s.Sector))
	}
}

// ceiling renders the ceiling portion of a slice.
func ceiling(c *state.Column) {
	mat := c.Sector.CeilSurface.Material
	extras := c.Sector.CeilSurface.ExtraStages
	transform := c.Sector.CeilSurface.Transform.Render
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
		*c.Sector.TopZ.Render - c.CameraZ}
	for c.ScreenY = c.EdgeTop; c.ScreenY < c.ClippedTop; c.ScreenY++ {
		c.RayFloorCeil[2] = float64(c.ScreenHeight/2 - c.ScreenY)
		screenIndex := uint32(c.ScreenX + c.ScreenY*c.ScreenWidth)
		denom := c.Sector.CeilNormal.Dot(&c.RayFloorCeil)
		if denom == 0 {
			c.FrameBuffer[screenIndex].AddPreMulColorSelf(&concepts.Vector4{1, 0, 0, 1})
			continue
		}

		t := planeRayDelta.Dot(&c.Sector.CeilNormal) / denom
		if t <= 0 {
			c.FrameBuffer[screenIndex].AddPreMulColorSelf(&concepts.Vector4{1, 1, 0, 1})
			dbg := fmt.Sprintf("%v ceiling t <= 0", c.Sector.Entity)
			c.DebugNotices.Push(dbg)
			continue
		}

		world[2] = c.RayFloorCeil[2] * t
		world[1] = c.RayFloorCeil[1] * t
		world[0] = c.RayFloorCeil[0] * t
		distToCeil := world.Length()
		dist2 := world.To2D().Length2()

		if distToCeil > c.ZBuffer[screenIndex] || dist2 > c.Distance*c.Distance {
			c.FrameBuffer[screenIndex].AddPreMulColorSelf(&concepts.Vector4{1, 0, 0, 1})
			continue
		}

		if distToCeil <= 0 {
			c.FrameBuffer[screenIndex].AddPreMulColorSelf(&concepts.Vector4{1, 0, 0, 1})
			continue
		}

		world[0] += c.Ray.Start[0]
		world[1] += c.Ray.Start[1]
		world[2] += c.CameraZ

		tx := (world[0] - sectorMin[0]) / (sectorMax[0] - sectorMin[0])
		ty := (world[1] - sectorMin[1]) / (sectorMax[1] - sectorMin[1])

		if mat != 0 {
			tx, ty = transform[0]*tx+transform[2]*ty+transform[4], transform[1]*tx+transform[3]*ty+transform[5]
			c.SampleShader(mat, extras, tx, ty, uint32(sw/distToCeil), uint32(sh/distToCeil))
			c.SampleLight(&c.MaterialSampler.Output, mat, &world, distToCeil)
		}
		//concepts.AsmVector4AddPreMulColorSelf((*[4]float64)(&s.FrameBuffer[screenIndex]), (*[4]float64)(&s.Material))
		c.FrameBuffer[screenIndex].AddPreMulColorSelf(&c.MaterialSampler.Output)
		c.ZBuffer[screenIndex] = distToCeil
	}
}
