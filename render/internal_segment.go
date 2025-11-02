// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package render

import (
	"tlyakhov/gofoom/components/selection"
	"tlyakhov/gofoom/concepts"
)

func (r *Renderer) renderInternalSegment(ewd *entityWithDistSq, block *block, xStart, xEnd int) {
	block.Sector = ewd.Sector
	block.IntersectedSegment = &ewd.InternalSegment.Segment
	block.IntersectionTop = ewd.InternalSegment.Top
	block.IntersectionBottom = ewd.InternalSegment.Bottom
	block.LightSampler.InputBody = 0
	block.LightSampler.Sector = block.Sector
	block.LightSampler.SegmentSector = block.Sector
	block.LightSampler.Segment = &ewd.InternalSegment.Segment
	ewd.InternalSegment.Normal.To3D(&block.LightSampler.Normal)

	for x := xStart; x < xEnd; x++ {
		block.Ray.Set(r.PlayerBody.Angle.Render*concepts.Deg2rad + r.ViewRadians[x])

		// Is the segment facing away?
		if !ewd.InternalSegment.TwoSided && block.Ray.Delta.Dot(&ewd.InternalSegment.Normal) > 0 {
			continue
		}
		// Ray intersection
		u := ewd.InternalSegment.Intersect2D(&block.Ray.Start, &block.Ray.End, &block.RaySegTest)
		if u < 0 {
			continue
		}
		block.ScreenX = x
		block.MaterialSampler.ScreenX = x
		block.MaterialSampler.Angle = block.Angle
		block.Distance = block.Ray.DistTo(&block.RaySegTest)
		block.RaySegIntersect[0] = block.RaySegTest[0]
		block.RaySegIntersect[1] = block.RaySegTest[1]
		block.segmentIntersection.U = u

		block.CalcScreen()

		if block.Pick && block.ScreenY >= block.ClippedTop && block.ScreenY <= block.ClippedBottom {
			vTop := float64(block.ScreenHeight/2) - block.ProjectedTop
			dv := (block.ProjectedTop - block.ProjectedBottom)
			if dv != 0 {
				dv = 1.0 / dv
			}
			v := (float64(block.ScreenY-int(block.ShearZ)) - vTop) * dv
			block.PickResult.Selection = append(block.PickResult.Selection, selection.SelectableFromInternalSegment(ewd.InternalSegment))
			block.PickResult.World[0] = block.RaySegTest[0]
			block.PickResult.World[1] = block.RaySegTest[1]
			block.PickResult.World[2] = block.IntersectionTop*(1.0-v) + v*block.IntersectionBottom
			block.PickResult.Normal[0] = ewd.InternalSegment.Normal[0]
			block.PickResult.Normal[1] = ewd.InternalSegment.Normal[1]
			block.PickResult.Normal[2] = 0
			return
		}
		r.wall(&block.column)
	}
}
