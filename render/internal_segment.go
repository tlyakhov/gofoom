// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package render

import (
	"tlyakhov/gofoom/components/selection"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/render/state"
)

func (r *Renderer) renderInternalSegment(ewd *state.EntityWithDist2, c *state.Column, xStart, xEnd int) {
	c.Sector = ewd.Sector
	for x := xStart; x < xEnd; x++ {
		c.Ray.Set(*r.PlayerBody.Angle.Render*concepts.Deg2rad + r.ViewRadians[x])

		// Is the segment facing away?
		if !ewd.InternalSegment.TwoSided && c.Ray.Delta.Dot(&ewd.InternalSegment.Normal) > 0 {
			continue
		}
		// Ray intersection
		if !ewd.InternalSegment.Intersect2D(&c.Ray.Start, &c.Ray.End, &c.RaySegTest) {
			continue
		}
		c.ScreenX = x
		c.MaterialSampler.ScreenX = x
		c.MaterialSampler.Angle = c.Angle
		c.Segment = &ewd.InternalSegment.Segment
		c.Distance = c.Ray.DistTo(&c.RaySegTest)
		c.RaySegIntersect[0] = c.RaySegTest[0]
		c.RaySegIntersect[1] = c.RaySegTest[1]
		c.SegmentIntersection.U = c.RaySegTest.Dist(ewd.InternalSegment.A) / ewd.InternalSegment.Length
		c.IntersectionTop = ewd.InternalSegment.Top
		c.IntersectionBottom = ewd.InternalSegment.Bottom
		c.CalcScreen()

		if c.Pick && c.ScreenY >= c.ClippedTop && c.ScreenY <= c.ClippedBottom {
			c.PickedSelection = append(c.PickedSelection, selection.SelectableFromInternalSegment(ewd.InternalSegment))
			return
		}
		c.LightSampler.Sector = c.Sector
		c.LightSampler.Segment = &ewd.InternalSegment.Segment
		ewd.InternalSegment.Normal.To3D(&c.LightSampler.Normal)
		r.wall(c)
	}
}
