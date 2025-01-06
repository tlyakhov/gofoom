// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package render

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/selection"
	"tlyakhov/gofoom/containers"
)

type block struct {
	column

	// Pre-allocated stack of past intersections, for speed
	Visited []segmentIntersection
	// Stack for walls to render over portals
	PortalWalls []*column
	// Maps for sorting bodies and internal segments
	Bodies           containers.Set[*core.Body]
	InternalSegments map[*core.InternalSegment]*core.Sector
	// For picking things in editor
	Pick            bool
	PickedSelection []*selection.Selectable
}
