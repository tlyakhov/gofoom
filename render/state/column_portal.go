// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package state

import (
	"math"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
)

type ColumnPortal struct {
	*Column
	Adj                                 *core.Sector
	AdjSegment                          *core.SectorSegment
	AdjTop, AdjBottom                   float64
	AdjProjectedTop, AdjProjectedBottom float64
	AdjScreenTop, AdjScreenBottom       int
	AdjClippedTop, AdjClippedBottom     int
}

func (cp *ColumnPortal) CalcScreen() {
	cp.Adj = core.SectorFromDb(cp.DB, cp.SectorSegment.AdjacentSector)
	cp.AdjSegment = cp.SectorSegment.AdjacentSegment
	cp.AdjBottom, cp.AdjTop = cp.Adj.SlopedZRender(cp.RaySegIntersect.To2D())
	cp.AdjProjectedTop = cp.ProjectZ(cp.AdjTop - cp.CameraZ)
	cp.AdjProjectedBottom = cp.ProjectZ(cp.AdjBottom - cp.CameraZ)
	cp.AdjScreenTop = cp.ScreenHeight/2 - int(math.Floor(cp.AdjProjectedTop))
	cp.AdjScreenBottom = cp.ScreenHeight/2 - int(math.Floor(cp.AdjProjectedBottom))
	cp.AdjClippedTop = concepts.Clamp(cp.AdjScreenTop, cp.ClippedTop, cp.ClippedBottom)
	cp.AdjClippedBottom = concepts.Clamp(cp.AdjScreenBottom, cp.ClippedTop, cp.ClippedBottom)
}
