// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package render

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/materials"
)

type entityWithDistSq struct {
	Body            *core.Body
	Visible         *materials.Visible
	InternalSegment *core.InternalSegment
	Sector          *core.Sector
	DistSq          float64
}
