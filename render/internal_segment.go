// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package render

import (
	"tlyakhov/gofoom/components/selection"
	"tlyakhov/gofoom/concepts"
)

func (r *Renderer) renderInternalSegment(ewd *entityWithDist2, c *column, xStart, xEnd int) {
	c.Sector = ewd.Sector
	c.Segment = &ewd.InternalSegment.Segment
	c.IntersectionTop = ewd.InternalSegment.Top
	c.IntersectionBottom = ewd.InternalSegment.Bottom
	c.LightSampler.InputBody = 0
	c.LightSampler.Sector = c.Sector
	c.LightSampler.Segment = &ewd.InternalSegment.Segment
	ewd.InternalSegment.Normal.To3D(&c.LightSampler.Normal)

	for x := xStart; x < xEnd; x++ {
		c.Ray.Set(*r.PlayerBody.Angle.Render*concepts.Deg2rad + r.ViewRadians[x])

		// Is the segment facing away?
		if !ewd.InternalSegment.TwoSided && c.Ray.Delta.Dot(&ewd.InternalSegment.Normal) > 0 {
			continue
		}
		// Ray intersection
		u := ewd.InternalSegment.Intersect2D(&c.Ray.Start, &c.Ray.End, &c.RaySegTest)
		if u < 0 {
			continue
		}
		c.ScreenX = x
		c.MaterialSampler.ScreenX = x
		c.MaterialSampler.Angle = c.Angle
		c.Distance = c.Ray.DistTo(&c.RaySegTest)
		c.RaySegIntersect[0] = c.RaySegTest[0]
		c.RaySegIntersect[1] = c.RaySegTest[1]
		c.segmentIntersection.U = u

		c.CalcScreen()

		if c.Pick && c.ScreenY >= c.ClippedTop && c.ScreenY <= c.ClippedBottom {
			c.PickedSelection = append(c.PickedSelection, selection.SelectableFromInternalSegment(ewd.InternalSegment))
			return
		}
		r.wall(c)
	}
}
