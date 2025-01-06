// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package render

import (
	"tlyakhov/gofoom/components/selection"
	"tlyakhov/gofoom/concepts"
)

func (r *Renderer) renderInternalSegment(ewd *entityWithDist2, block *block, xStart, xEnd int) {
	block.Sector = ewd.Sector
	block.Segment = &ewd.InternalSegment.Segment
	block.IntersectionTop = ewd.InternalSegment.Top
	block.IntersectionBottom = ewd.InternalSegment.Bottom
	block.LightSampler.InputBody = 0
	block.LightSampler.Sector = block.Sector
	block.LightSampler.Segment = &ewd.InternalSegment.Segment
	ewd.InternalSegment.Normal.To3D(&block.LightSampler.Normal)

	for x := xStart; x < xEnd; x++ {
		block.Ray.Set(*r.PlayerBody.Angle.Render*concepts.Deg2rad + r.ViewRadians[x])

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
			block.PickedSelection = append(block.PickedSelection, selection.SelectableFromInternalSegment(ewd.InternalSegment))
			return
		}
		r.wall(&block.column)
	}
}
