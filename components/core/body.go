// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package core

import (
	"math"

	"tlyakhov/gofoom/dynamic"
	"tlyakhov/gofoom/ecs"

	"tlyakhov/gofoom/concepts"

	"github.com/spf13/cast"
)

//go:generate gofoom_ecs_generator $GOFILE
type Body struct {
	ecs.Attached `editable:"^"`
	Pos          dynamic.DynamicValue[concepts.Vector3] `editable:"Position"`
	Size         dynamic.DynamicValue[concepts.Vector2] `editable:"Size"`
	Angle        dynamic.DynamicValue[float64]          `editable:"Angle"`

	StepSound ecs.Entity `editable:"Step Sound" edit_type:"Sound"`

	SectorEntity ecs.Entity
	QuadNode     *QuadNode `ecs:"norelation"`
	OnGround     bool
}

func (b *Body) String() string {
	return "Body"
}

func (b *Body) OnDetach(e ecs.Entity) {
	defer b.Attached.OnDetach(e)
	if !b.IsAttached() {
		return
	}
	if sector := b.Sector(); sector != nil {
		delete(sector.Bodies, e)
	}

	if b.QuadNode != nil {
		b.QuadNode.Remove(b)
		b.QuadNode = nil
	}
}

func (b *Body) OnDelete() {
	defer b.Attached.OnDelete()
	if b.IsAttached() {
		b.Pos.Detach(ecs.Simulation)
		b.Size.Detach(ecs.Simulation)
		b.Angle.Detach(ecs.Simulation)
	}
}

func (b *Body) OnAttach() {
	b.Attached.OnAttach()
	b.Pos.Attach(ecs.Simulation)
	b.Size.Attach(ecs.Simulation)
	b.Angle.Attach(ecs.Simulation)

	if tree := ecs.Singleton(QuadtreeCID).(*Quadtree); tree != nil {
		tree.Update(b)
	}
}

func (b *Body) Sector() *Sector {
	return GetSector(b.SectorEntity)
}

func (b *Body) Normal() *concepts.Vector2 {
	dy, dx := math.Sincos(b.Angle.Render * concepts.Deg2rad)
	return &concepts.Vector2{dx, dy}
}

func (b *Body) Angle2DTo(p *concepts.Vector3) float64 {
	dx := b.Pos.Now[0] - p[0]
	dy := b.Pos.Now[1] - p[1]
	return math.Atan2(dy, dx)*concepts.Rad2deg + 180.0
}

func (b *Body) RenderSector() *Sector {
	// Figure out the sector to use for rendering this body based on the current
	// render position. This may be different from Body.Sector(), since
	// the Render position is interpolated (see Simulation)
	// First, fast path - check if the body is inside its .Sector()
	p := b.Pos.Render.To2D()
	sector := b.Sector()
	if sector != nil && sector.IsPointInside2D(p) {
		return sector
	}
	// Go through all sectors to find the containing one. Optimize this later if
	// necessary.
	arena := ecs.ArenaFor[Sector](SectorCID)
	for i := range arena.Cap() {
		if sector := arena.Value(i); sector != nil && sector.IsPointInside2D(p) {
			return sector
		}
	}
	return nil
}

var defaultBodySize = "10,10"

func (b *Body) Construct(data map[string]any) {
	b.Attached.Construct(data)

	b.Pos.Construct(nil)
	b.Size.Construct(defaultBodySize)
	b.Angle.Construct(nil)

	b.Angle.IsAngle = true

	if data == nil {
		return
	}

	if v, ok := data["Pos"]; ok {
		b.Pos.Construct(v)
	}

	if v, ok := data["Size"]; ok {
		b.Size.Construct(v)
	}
	if v, ok := data["Angle"]; ok {
		b.Angle.Construct(v)
	}
	if v, ok := data["StepSound"]; ok {
		b.StepSound, _ = ecs.ParseEntity(cast.ToString(v))
	}
}

func (b *Body) Serialize() map[string]any {
	result := b.Attached.Serialize()
	result["Pos"] = b.Pos.Serialize()
	result["Size"] = b.Size.Serialize()
	result["Angle"] = b.Angle.Serialize()
	if b.StepSound != 0 {
		result["StepSound"] = b.StepSound.Serialize()
	}
	return result
}
