// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package render

import (
	"math"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/dynamic"
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
	cp.Adj = core.GetSector(cp.ECS, cp.SectorSegment.AdjacentSector)
	cp.AdjSegment = cp.SectorSegment.AdjacentSegment
	cp.AdjBottom, cp.AdjTop = cp.Adj.ZAt(dynamic.DynamicRender, cp.RaySegIntersect.To2D())
	cp.AdjProjectedTop = cp.ProjectZ(cp.AdjTop - cp.CameraZ)
	cp.AdjProjectedBottom = cp.ProjectZ(cp.AdjBottom - cp.CameraZ)

	adjScreenTop := cp.ScreenHeight/2 - int(math.Floor(cp.AdjProjectedTop))
	adjScreenBottom := cp.ScreenHeight/2 - int(math.Floor(cp.AdjProjectedBottom))
	cp.AdjClippedTop = concepts.Clamp(adjScreenTop, cp.ClippedTop, cp.ClippedBottom)
	cp.AdjClippedBottom = concepts.Clamp(adjScreenBottom, cp.ClippedTop, cp.ClippedBottom)
}
