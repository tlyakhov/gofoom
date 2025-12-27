// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"log"
	"math"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/controllers"
	"tlyakhov/gofoom/ecs"
	"tlyakhov/gofoom/editor/state"

	"fyne.io/fyne/v2/driver/desktop"
)

// Declare conformity with interfaces
var _ Placeable = (*AddSector)(nil)

type AddSector struct {
	AddEntity
	Sector *core.Sector
}

func (a *AddSector) Activate() {
	a.AddEntity.Activate()
	a.Sector = a.Components.Get(core.SectorCID).(*core.Sector)
	a.SetMapCursor(desktop.CrosshairCursor)
}

func (a *AddSector) newSegment() {
	seg := core.SectorSegment{}
	seg.Construct(nil)
	seg.Sector = a.Sector
	seg.HiSurface.Material = controllers.DefaultMaterial()
	seg.LoSurface.Material = controllers.DefaultMaterial()
	seg.Surface.Material = controllers.DefaultMaterial()
	seg.P.SetAll(*a.WorldGrid(&a.State().MouseDownWorld))

	segs := a.Sector.Segments
	if len(segs) > 0 {
		seg.Prev = segs[len(segs)-1]
		seg.Next = segs[0]
		seg.Next.Prev = &seg
		seg.Prev.Next = &seg
	}

	a.Sector.Segments = append(a.Sector.Segments, &seg)
	a.Sector.Precompute()
}

func (a *AddSector) Point() bool {
	if !a.AddEntity.Point() {
		return false
	}
	a.State().Lock.Lock()
	defer a.State().Lock.Unlock()

	if len(a.Sector.Segments) == 0 {
		a.newSegment()
	}
	segs := a.Sector.Segments
	seg := segs[len(segs)-1]
	worldGrid := a.WorldGrid(&a.State().MouseWorld)
	seg.P.SetAll(*worldGrid)
	seg.Precompute()
	if len(segs) > 1 {
		seg.Prev.Precompute()
	}
	return true
}

func (a *AddSector) guessLayer() {
	arena := ecs.ArenaFor[core.Sector](core.SectorCID)
	highestLayer := math.MinInt32
	for i := range arena.Cap() {
		sector := arena.Value(i)
		if sector == nil {
			continue
		}
		if sector.Layer > highestLayer && sector.Contains2D(a.Sector) {
			highestLayer = sector.Layer
		}
	}
	if highestLayer != math.MinInt32 {
		a.Sector.Layer = highestLayer + 1
	}
}

func (a *AddSector) EndPoint() bool {
	log.Printf("Tried to EndPoint on AddSector: %v", a.Sector.String())
	if a.Mode != "Placing" {
		return false
	}
	segs := a.Sector.Segments
	if len(segs) > 1 {
		first := segs[0]
		last := segs[len(segs)-1]
		if last.P.Render.Sub(&first.P.Render).Length() < state.SegmentSelectionEpsilon {
			a.State().Lock.Lock()
			a.Sector.Segments = segs[:(len(segs) - 1)]
			a.Sector.Precompute()
			a.Sector.TransformOrigin = *a.Sector.Center.Spawn.To2D()
			a.guessLayer()
			a.State().Lock.Unlock()
			return a.AddEntity.EndPoint()
		}
	}
	a.newSegment()
	return true
}

func (a *AddSector) Status() string {
	return "Click to place " + a.Mode
}
