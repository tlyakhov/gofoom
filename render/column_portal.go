// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package render

import (
	"math"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
)

type columnPortal struct {
	*block
	Adj                                 *core.Sector
	AdjSegment                          *core.SectorSegment
	AdjTop, AdjBottom                   float64
	AdjProjectedTop, AdjProjectedBottom float64
	AdjClippedTop, AdjClippedBottom     int
}

func (cp *columnPortal) CalcScreen() {
	if cp.Adj.Top.Ignore {
		cp.AdjTop = cp.IntersectionTop
	} else {
		cp.AdjTop = cp.Adj.Top.ZAt(cp.RaySegIntersect.To2D())
		cp.TopPlane = &cp.Adj.Top
	}
	if cp.Adj.Bottom.Ignore {
		cp.AdjBottom = cp.IntersectionBottom
	} else {
		cp.AdjBottom = cp.Adj.Bottom.ZAt(cp.RaySegIntersect.To2D())
		cp.BottomPlane = &cp.Adj.Bottom
	}

	cp.AdjProjectedTop = cp.ProjectZ(cp.AdjTop - cp.CameraZ)
	cp.AdjProjectedBottom = cp.ProjectZ(cp.AdjBottom - cp.CameraZ)

	adjScreenTop := cp.ScreenHeight/2 - int(math.Floor(cp.AdjProjectedTop)) + int(cp.ShearZ)
	adjScreenBottom := cp.ScreenHeight/2 - int(math.Floor(cp.AdjProjectedBottom)) + int(cp.ShearZ)
	cp.AdjClippedTop = concepts.Clamp(adjScreenTop, cp.ClippedTop, cp.ClippedBottom)
	cp.AdjClippedBottom = concepts.Clamp(adjScreenBottom, cp.ClippedTop, cp.ClippedBottom)
}
