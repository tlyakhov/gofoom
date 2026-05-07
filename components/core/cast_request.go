// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package core

import (
	"tlyakhov/gofoom/concepts"
)

type CastResponse struct {
	HitSegment *SectorSegment
	HitPoint   concepts.Vector3
	// You can also pass in a max limit in this field
	HitDistSq  float64
	HitPortal  int // -1 = lo, 0 = mid, 1 = hi
	NextSector *Sector
}

type CastRequest struct {
	// Inputs
	*concepts.Ray

	IgnoreSegment *Segment
	MinDistSq     float64
	CheckEntry    bool
	Debug         bool
	IgnoreZ       bool

	// Input/Output
	CastResponse
}
