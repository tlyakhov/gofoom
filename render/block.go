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

	// Stack for walls to render over portals
	PortalWalls []*column
	// Maps for sorting bodies and internal segments
	Bodies           containers.Set[*core.Body]
	InternalSegments map[*core.InternalSegment]*core.Sector
	// For picking things in editor
	Pick            bool
	PickedSelection []*selection.Selectable
}

func (b *block) teleportRay() {
	b.SectorSegment.PortalMatrix.UnprojectSelf(&b.Ray.Start)
	b.SectorSegment.PortalMatrix.UnprojectSelf(&b.Ray.End)
	b.SectorSegment.AdjacentSegment.MirrorPortalMatrix.ProjectSelf(&b.Ray.Start)
	b.SectorSegment.AdjacentSegment.MirrorPortalMatrix.ProjectSelf(&b.Ray.End)
	b.Ray.AnglesFromStartEnd()
	// TODO: this has a bug if the adjacent sector has a sloped floor.
	// Getting the right floor height is a bit expensive because we have to
	// project the intersection point. For now just use the sector minimum.
	b.CameraZ = b.CameraZ - b.IntersectionBottom + b.SectorSegment.AdjacentSegment.Sector.Min[2]
	b.RayPlane[0] = b.Ray.AngleCos * b.ViewFix[b.ScreenX]
	b.RayPlane[1] = b.Ray.AngleSin * b.ViewFix[b.ScreenX]
	b.MaterialSampler.Ray = &b.Ray
}
