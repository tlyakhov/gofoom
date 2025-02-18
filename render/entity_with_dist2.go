// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package render

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/materials"
)

type entityWithDist2 struct {
	Body            *core.Body
	Visible         *materials.Visible
	InternalSegment *core.InternalSegment
	Sector          *core.Sector
	Dist2           float64
}
