// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/ecs"
)

// TODO: This can be very expensive for large areas with lots of lights. How can
// we optimize this?
func updatePVS(pvsSector *core.Sector, normals []*concepts.Vector2, visitor *core.Sector, min, max *concepts.Vector3) {
	if visitor == nil {
		pvsSector.PVS = make(map[ecs.Entity]*core.Sector)
		pvsSector.PVL = make([]*core.Body, 0)
		pvsSector.Colliders = make(map[ecs.Entity]*core.Mobile)
		pvsSector.PVS[pvsSector.Entity] = pvsSector
		visitor = pvsSector
	}

	for entity, body := range visitor.Bodies {
		if core.GetLight(body.ECS, entity) != nil {
			pvsSector.PVL = append(pvsSector.PVL, body)
		}
		if m := core.GetMobile(body.ECS, entity); m != nil &&
			(m.CrBody != core.CollideNone || m.CrPlayer != core.CollideNone) {
			pvsSector.Colliders[entity] = m
		}
	}

	if min == nil || max == nil {
		min, max = &pvsSector.Min, &pvsSector.Max
	}
	nNormals := len(normals)
	normals = append(normals, nil)

	for _, seg := range visitor.Segments {
		adj := seg.AdjacentSegment
		if adj == nil {
			continue
		}
		correctSide := true
		for _, normal := range normals[:nNormals] {
			correctSide = correctSide && normal.Dot(&seg.Normal) >= 0
		}
		if !correctSide || pvsSector.PVS[adj.Sector.Entity] != nil {
			continue
		}
		if adj.Sector.Min[2] >= max[2] || adj.Sector.Max[2] <= min[2] {
			continue
		}
		adjmax := max
		adjmin := min
		if adj.Sector.Max[2] < max[2] {
			adjmax = &adj.Sector.Max
		}
		if adj.Sector.Min[2] > min[2] {
			adjmin = &adj.Sector.Min
		}

		adjsec := core.GetSector(pvsSector.ECS, seg.AdjacentSector)
		pvsSector.PVS[seg.AdjacentSector] = adjsec

		normals[nNormals] = &seg.Normal
		updatePVS(pvsSector, normals, adjsec, adjmin, adjmax)
	}
}
