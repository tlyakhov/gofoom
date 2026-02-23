// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/components/selection"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/ecs"
)

type MarkMakerController struct {
	ecs.BaseController
	*materials.MarkMaker
	pos       *concepts.Vector3
	transform concepts.Matrix2
}

func init() {
	ecs.Types().RegisterController(func() ecs.Controller { return &MarkMakerController{} }, 100)
}

func (mmc *MarkMakerController) ComponentID() ecs.ComponentID {
	return materials.MarkMakerCID
}

func (mmc *MarkMakerController) Methods() ecs.ControllerMethod {
	return ecs.ControllerPrecompute
}

func (mmc *MarkMakerController) Target(target ecs.Component, e ecs.Entity) bool {
	mmc.Entity = e
	mmc.MarkMaker = target.(*materials.MarkMaker)
	return mmc.MarkMaker != nil && mmc.MarkMaker.IsActive()
}

func (mmc *MarkMakerController) Precompute() {
}

func (mmc *MarkMakerController) addMark(mark materials.Mark) {
	mmc.Marks.PushBack(mark)
	for mmc.Marks.Len() > constants.MaxWeaponMarks {
		mark := mmc.Marks.PopFront()
		for i, stage := range mark.Surface.ExtraStages {
			if stage != mark.ShaderStage {
				continue
			}
			mark.Surface.ExtraStages = append(mark.Surface.ExtraStages[:i], mark.Surface.ExtraStages[i+1:]...)
			break
		}
	}
}

func (mmc *MarkMakerController) surfaceFromSelectable(s *selection.Selectable, p *concepts.Vector2) (surf *materials.Surface, bottom, top float64) {
	var adj *core.Sector
	var adjSegment *core.SectorSegment

	// Will be invalid for InternalSegments
	if s.SectorSegment != nil {
		adjSegment = s.SectorSegment.AdjacentSegment
	}
	// Handle surfaces of edges between different layers
	if adjSegment == nil &&
		(s.Type == selection.SelectableHi || s.Type == selection.SelectableLow) {
		lower := s.Sector.OverlapAt(p, true)
		if lower == nil {
			return nil, 0, 0
		}
		adj = lower
		adjSegment = s.SectorSegment
	} else if adjSegment != nil {
		adj = adjSegment.Sector
	}

	switch s.Type {
	case selection.SelectableHi:
		top = s.Sector.Top.ZAt(p)
		adjTop := adj.Top.ZAt(p)
		if adjTop <= top {
			bottom = adjTop
			surf = &adjSegment.HiSurface
		} else {
			bottom, top = top, adjTop
			surf = &s.SectorSegment.HiSurface
		}
	case selection.SelectableLow:
		bottom = s.Sector.Bottom.ZAt(p)
		adjBottom := adj.Bottom.ZAt(p)
		if bottom <= adjBottom {
			top = adjBottom
			surf = &adjSegment.LoSurface
		} else {
			bottom, top = adjBottom, bottom
			surf = &s.SectorSegment.LoSurface
		}
	case selection.SelectableMid:
		bottom, top = s.Sector.ZAt(p)
		surf = &s.SectorSegment.Surface
	case selection.SelectableInternalSegment:
		bottom, top = s.InternalSegment.Bottom, s.InternalSegment.Top
		surf = &s.InternalSegment.Surface
	}
	return
}

// TODO: This is more generally useful as a way to create transforms
// that map world-space onto texture space. This should be refactored to be part
// of anything with a surface
func (mmc *MarkMakerController) markSurfaceAndTransform(s *selection.Selectable, transform *concepts.Matrix2) *materials.Surface {
	// Inverse of the size of bullet mark we want
	scale := 1.0 / mmc.Size
	hit2d := mmc.pos.To2D()
	// 3x2 transformation matrixes are composed of
	// the horizontal basis vector in slots [0] & [1], which we set to the
	// width of the segment, scaled
	// the vertical basis vector in slots [2] & [3], which we set to the
	// height of the segment
	// and finally the translation in slots [4] & [5], which we set to the
	// world position of the mark, relative to the segment
	switch s.Type {
	case selection.SelectableHi, selection.SelectableLow, selection.SelectableMid:
		transform[concepts.MatBasis1X] = s.SectorSegment.Length
		transform[concepts.MatTransX] = -hit2d.Dist(&s.SectorSegment.P.Render)
	case selection.SelectableInternalSegment:
		transform[concepts.MatBasis1X] = s.InternalSegment.Length
		transform[concepts.MatTransX] = -hit2d.Dist(s.InternalSegment.A)
	}

	surf, bottom, top := mmc.surfaceFromSelectable(s, hit2d)
	if surf == nil {
		return nil
	}
	// This is reversed because our UV coordinates go top->bottom
	transform[concepts.MatBasis2Y] = (bottom - top)
	transform[concepts.MatTransY] = -(mmc.pos[2] - top)

	transform[concepts.MatBasis1X] *= scale
	transform[concepts.MatBasis2Y] *= scale
	transform[concepts.MatTransX] *= scale
	transform[concepts.MatTransY] *= scale

	return surf
}

func (mmc *MarkMakerController) MakeMark(s *selection.Selectable, pos *concepts.Vector3) {
	mmc.pos = pos
	switch s.Type {
	case selection.SelectableSectorSegment, selection.SelectableHi,
		selection.SelectableLow, selection.SelectableMid,
		selection.SelectableInternalSegment:
		// Make a mark on walls

		// TODO: Include floors and ceilings
		shaderStage := &materials.ShaderStage{
			Material:               mmc.Material,
			IgnoreSurfaceTransform: false,
			Tag:                    "MarkMaker " + mmc.Entity.Serialize(),
		}
		// TODO: Fix this
		//es.CFlags = ecs.ComponentInternal
		shaderStage.Construct(nil)
		shaderStage.Flags = 0
		surf := mmc.markSurfaceAndTransform(s, &mmc.transform)
		if surf == nil {
			return
		}
		surf.ExtraStages = append(surf.ExtraStages, shaderStage)
		shaderStage.Transform.From(&surf.Transform.Now)
		shaderStage.Transform.AffineInverseSelf().MulSelf(&mmc.transform)
		mmc.addMark(materials.Mark{
			ShaderStage: shaderStage,
			Surface:     surf,
		})
	}
}
