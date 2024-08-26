// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package core

import (
	"math"

	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/ecs"

	"tlyakhov/gofoom/concepts"
)

type Body struct {
	ecs.Attached `editable:"^"`
	Pos          ecs.DynamicValue[concepts.Vector3] `editable:"Position"`
	Vel          ecs.DynamicValue[concepts.Vector3]
	Force        concepts.Vector3
	Size         ecs.DynamicValue[concepts.Vector2] `editable:"Size"`
	SectorEntity ecs.Entity
	Angle        ecs.DynamicValue[float64] `editable:"Angle"`
	Mass         float64                   `editable:"Mass"`

	// "Bounciness" (0 = inelastic, 1 = perfectly elastic)
	Elasticity  float64           `editable:"Elasticity"`
	CrBody      CollisionResponse `editable:"Collision (Body)"`
	CrPlayer    CollisionResponse `editable:"Collision (Player)"`
	CrWall      CollisionResponse `editable:"Collision (Wall)"`
	Shadow      BodyShadow        `editable:"Shadow Type"`
	MountHeight float64           `editable:"Mount Height"`
	OnGround    bool

	// For rendering - we can figure out which side of this our render
	// position is on so we end up in the right sector when render blending
	LastEnteredPortal *SectorSegment
}

var BodyComponentIndex int

func init() {
	BodyComponentIndex = ecs.RegisterComponent(&ecs.Column[Body, *Body]{Getter: GetBody})
}

func GetBody(db *ecs.ECS, e ecs.Entity) *Body {
	if asserted, ok := db.Component(e, BodyComponentIndex).(*Body); ok {
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
		b.Vel.Detach(b.ECS.Simulation)
		b.Size.Detach(b.ECS.Simulation)
		b.Angle.Detach(b.ECS.Simulation)
	}
	if sector := b.Sector(); sector != nil {
		delete(sector.Bodies, b.Entity)
	}
	b.Attached.OnDetach()
}

func (b *Body) AttachECS(db *ecs.ECS) {
	if b.ECS != db {
		b.OnDetach()
	}
	b.Attached.AttachECS(db)
	b.Pos.Attach(db.Simulation)
	b.Vel.Attach(db.Simulation)
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

func (b *Body) BillboardSegment(unitView *concepts.Vector3, ds ecs.DynamicStage) *Segment {
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
	col := ecs.ColumnFor[Sector](b.ECS, SectorComponentIndex)
	for i := range col.Length {
		sector := col.Value(i)
		if sector.Entity == 0 || !sector.IsPointInside2D(p) {
			continue
		}
		return sector
	}
	return nil
}

var defaultBodySize = map[string]any{"Original": map[string]any{"X": 10.0, "Y": 10.0}}

func (b *Body) Construct(data map[string]any) {
	b.Attached.Construct(data)

	b.Pos.Construct(nil)
	b.Vel.Construct(nil)
	b.Size.Construct(defaultBodySize)
	b.Angle.Construct(nil)

	b.Elasticity = 0.5
	b.CrBody = CollideNone
	b.CrPlayer = CollideNone
	b.CrWall = CollideSeparate
	b.MountHeight = constants.PlayerMountHeight
	b.Shadow = BodyShadowNone
	b.Angle.NoRenderBlend = true

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
	if v, ok := data["Vel"]; ok {
		v3 := v.(map[string]any)
		if _, ok2 := v3["X"]; ok2 {
			v3 = map[string]any{"Original": v3}
		}
		b.Vel.Construct(v3)
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
	if v, ok := data["MountHeight"]; ok {
		b.MountHeight = v.(float64)
	}
	if v, ok := data["Mass"]; ok {
		b.Mass = v.(float64)
	}
	if v, ok := data["Elasticity"]; ok {
		b.Elasticity = v.(float64)
	}
	if v, ok := data["CrBody"]; ok {
		c, err := CollisionResponseString(v.(string))
		if err == nil {
			b.CrBody = c
		}
	}
	if v, ok := data["CrPlayer"]; ok {
		c, err := CollisionResponseString(v.(string))
		if err == nil {
			b.CrPlayer = c
		}
	}
	if v, ok := data["CrWall"]; ok {
		c, err := CollisionResponseString(v.(string))
		if err == nil {
			b.CrWall = c
		}
	}
	if v, ok := data["Shadow"]; ok {
		c, err := BodyShadowString(v.(string))
		if err == nil {
			b.Shadow = c
		} else {
			panic(err)
		}
	}
}

func (b *Body) Serialize() map[string]any {
	result := b.Attached.Serialize()
	result["Pos"] = b.Pos.Serialize()
	result["Vel"] = b.Vel.Serialize()
	result["Size"] = b.Size.Serialize()
	result["Angle"] = b.Angle.Serialize()
	result["Mass"] = b.Mass
	result["Elasticity"] = b.Elasticity
	result["MountHeight"] = b.MountHeight
	result["CrBody"] = b.CrBody.String()
	result["CrPlayer"] = b.CrPlayer.String()
	result["CrWall"] = b.CrWall.String()
	if b.Shadow != BodyShadowNone {
		result["Shadow"] = b.Shadow.String()
	}
	return result
}
