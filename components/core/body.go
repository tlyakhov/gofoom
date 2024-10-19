// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package core

import (
	"math"

	"tlyakhov/gofoom/dynamic"
	"tlyakhov/gofoom/ecs"

	"tlyakhov/gofoom/concepts"
)

type Body struct {
	ecs.Attached `editable:"^"`
	Pos          dynamic.DynamicValue[concepts.Vector3] `editable:"Position"`
	Size         dynamic.DynamicValue[concepts.Vector2] `editable:"Size"`
	Angle        dynamic.DynamicValue[float64]          `editable:"Angle"`

	SectorEntity ecs.Entity
	OnGround     bool
}

var BodyCID ecs.ComponentID

func init() {
	BodyCID = ecs.RegisterComponent(&ecs.Column[Body, *Body]{Getter: GetBody}, "BodySectorSegment")
}

func GetBody(db *ecs.ECS, e ecs.Entity) *Body {
	if asserted, ok := db.Component(e, BodyCID).(*Body); ok {
		return asserted
	}
	return nil
}

func (b *Body) String() string {
	return "Body: " + b.Pos.Now.StringHuman()
}

func (b *Body) OnDetach() {
	if b.ECS != nil {
		b.Pos.Detach(b.ECS.Simulation)
		b.Size.Detach(b.ECS.Simulation)
		b.Angle.Detach(b.ECS.Simulation)
		if sector := b.Sector(); sector != nil {
			delete(sector.Bodies, b.Entity)
		}
	}
	b.Attached.OnDetach()
}

func (b *Body) AttachECS(db *ecs.ECS) {
	if b.ECS != db {
		b.OnDetach()
	}
	b.Attached.AttachECS(db)
	b.Pos.Attach(db.Simulation)
	b.Size.Attach(db.Simulation)
	b.Angle.Attach(b.ECS.Simulation)
}

func (b *Body) Sector() *Sector {
	return GetSector(b.ECS, b.SectorEntity)
}

func (b *Body) Normal() *concepts.Vector2 {
	dy, dx := math.Sincos(*b.Angle.Render * concepts.Deg2rad)
	return &concepts.Vector2{dx, dy}
}

func (b *Body) Angle2DTo(p *concepts.Vector3) float64 {
	dx := b.Pos.Now[0] - p[0]
	dy := b.Pos.Now[1] - p[1]
	return math.Atan2(dy, dx)*concepts.Rad2deg + 180.0
}

func (b *Body) BillboardSegment(unitView *concepts.Vector3, ds dynamic.DynamicStage) *Segment {
	p := b.Pos.Value(ds)
	s := b.Size.Value(ds)
	return &Segment{
		A: &concepts.Vector2{
			p[0] + unitView[1]*s[0]*0.5,
			p[1] - unitView[0]*s[0]*0.5,
		},
		B: &concepts.Vector2{
			p[0] - unitView[1]*s[0]*0.5,
			p[1] + unitView[0]*s[0]*0.5,
		}}
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
	col := ecs.ColumnFor[Sector](b.ECS, SectorCID)
	for i := range col.Cap() {
		if sector := col.Value(i); sector != nil && sector.IsPointInside2D(p) {
			return sector
		}
	}
	return nil
}

var defaultBodySize = map[string]any{"Original": map[string]any{"X": 10.0, "Y": 10.0}}

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
		v3 := v.(map[string]any)
		if _, ok2 := v3["X"]; ok2 {
			v3 = map[string]any{"Original": v3}
		}
		b.Pos.Construct(v3)
	}

	if v, ok := data["Size"]; ok {
		v2 := v.(map[string]any)
		if _, ok2 := v2["X"]; ok2 {
			v2 = map[string]any{"Original": v2}
		}
		b.Size.Construct(v2)
	}
	if v, ok := data["Angle"]; ok {
		if v2, ok2 := v.(float64); ok2 {
			v = map[string]any{"Original": v2}
		}
		b.Angle.Construct(v.(map[string]any))
	}
}

func (b *Body) Serialize() map[string]any {
	result := b.Attached.Serialize()
	result["Pos"] = b.Pos.Serialize()
	result["Size"] = b.Size.Serialize()
	result["Angle"] = b.Angle.Serialize()
	return result
}
