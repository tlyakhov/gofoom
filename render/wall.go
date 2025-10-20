// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package render

import (
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/components/selection"
	"tlyakhov/gofoom/concepts"
)

func wallPick(b *block) {
	if b.ScreenY < b.ClippedTop || b.ScreenY >= b.ClippedBottom {
		return
	}
	vTop := float64(b.ScreenHeight/2) - b.ProjectedTop
	dv := (b.ProjectedTop - b.ProjectedBottom)
	if dv != 0 {
		dv = 1.0 / dv
	}
	v := (float64(b.ScreenY-int(b.ShearZ)) - vTop) * dv
	b.PickResult.Selection = append(b.PickResult.Selection, selection.SelectableFromWall(b.IntersectedSectorSegment, selection.SelectableMid))
	b.PickResult.World[0] = b.RaySegIntersect[0]
	b.PickResult.World[1] = b.RaySegIntersect[1]
	b.PickResult.World[2] = b.IntersectionTop*(1.0-v) + v*b.IntersectionBottom
	b.PickResult.Normal[0] = b.IntersectedSectorSegment.Normal[0]
	b.PickResult.Normal[1] = b.IntersectedSectorSegment.Normal[1]
	b.PickResult.Normal[2] = 0

}

// wall renders the wall portion (potentially over a portal).
func (r *Renderer) wall(c *column) {
	mat := c.IntersectedSegment.Surface.Material
	lit := materials.GetLit(c.IntersectedSegment.Surface.Material)
	extras := c.IntersectedSegment.Surface.ExtraStages
	c.MaterialSampler.Initialize(mat, extras)
	transform := c.IntersectedSegment.Surface.Transform.Render
	noSlope := c.IntersectedSectorSegment != nil && c.IntersectedSectorSegment.WallUVIgnoreSlope
	// To calculate the vertical texture coordinate, we can't use the integer
	// screen coordinates, we need to use the precise floats
	dv := 0.0
	vTop := float64(c.ScreenHeight / 2)
	if noSlope {
		vTop -= c.ProjectedSectorTop
		dv = c.ProjectedSectorTop - c.ProjectedSectorBottom
	} else {
		vTop -= c.ProjectedTop
		dv = (c.ProjectedTop - c.ProjectedBottom)
	}
	c.ScaleW = uint32(c.ProjectZ(c.IntersectedSegment.Length))
	c.ScaleH = uint32(dv)
	if dv != 0 {
		dv = 1.0 / dv
	}

	for c.ScreenY = c.ClippedTop; c.ScreenY < c.ClippedBottom; c.ScreenY++ {
		screenIndex := uint32(c.ScreenX + c.ScreenY*c.ScreenWidth)

		if c.Distance >= c.ZBuffer[screenIndex] {
			continue
		}

		v := (float64(c.ScreenY-int(c.ShearZ)) - vTop) * dv
		c.RaySegIntersect[2] = c.IntersectionTop*(1.0-v) + v*c.IntersectionBottom

		if mat != 0 {
			c.MaterialSampler.NU = c.segmentIntersection.U
			c.MaterialSampler.NV = v
			c.MaterialSampler.U = transform[0]*c.MaterialSampler.NU + transform[2]*c.MaterialSampler.NV + transform[4]
			c.MaterialSampler.V = transform[1]*c.MaterialSampler.NU + transform[3]*c.MaterialSampler.NV + transform[5]
			c.SampleMaterial(extras)
			if lit != nil {
				c.SampleLight(&c.MaterialSampler.Output, lit, &c.RaySegIntersect, c.Distance)
			}
		}
		concepts.BlendColors(&r.FrameBuffer[screenIndex], &c.MaterialSampler.Output, 1.0)
		if c.MaterialSampler.Output[3] > 0.8 {
			r.ZBuffer[screenIndex] = c.Distance
		}
	}
}
