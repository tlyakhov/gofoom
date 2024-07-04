// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package core

import (
	"math"

	"tlyakhov/gofoom/constants"

	"tlyakhov/gofoom/concepts"
)

type Body struct {
	concepts.Attached `editable:"^"`
	Pos               concepts.SimVariable[concepts.Vector3] `editable:"Position"`
	Vel               concepts.SimVariable[concepts.Vector3]
	Force             concepts.Vector3
	Size              concepts.SimVariable[concepts.Vector2] `editable:"Size"`
	SectorEntity      concepts.Entity
	Angle             concepts.SimVariable[float64] `editable:"Angle"`
	Mass              float64                       `editable:"Mass"`
	CollisionResponse CollisionResponse             `editable:"Collision Response"`
	Shadow            BodyShadow                    `editable:"Shadow Type"`
	MountHeight       float64                       `editable:"Mount Height"`
	OnGround          bool

	// For rendering - we can figure out which side of this our render
	// position is on so we end up in the right sector when render blending
	LastEnteredPortal *SectorSegment
}

var BodyComponentIndex int

func init() {
	BodyComponentIndex = concepts.DbTypes().Register(Body{}, BodyFromDb)
}

func BodyFromDb(db *concepts.EntityComponentDB, e concepts.Entity) *Body {
	if asserted, ok := db.Component(e, BodyComponentIndex).(*Body); ok {
		return asserted
	}
	return nil
}

func (b *Body) String() string {
	return "Body: " + b.Pos.Now.StringHuman()
}

func (b *Body) SetDB(db *concepts.EntityComponentDB) {
	if b.DB != nil {
		b.Pos.Detach(b.DB.Simulation)
		b.Vel.Detach(b.DB.Simulation)
		b.Size.Detach(b.DB.Simulation)
		b.Angle.Detach(b.DB.Simulation)
	}
	b.Attached.SetDB(db)
	b.Pos.Attach(db.Simulation)
	b.Vel.Attach(db.Simulation)
	b.Size.Attach(db.Simulation)
	b.Angle.Attach(b.DB.Simulation)
}

func (b *Body) Sector() *Sector {
	return SectorFromDb(b.DB, b.SectorEntity)
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
	for _, a := range b.DB.Components[SectorComponentIndex] {
		sector = a.(*Sector)
		if sector == nil || !sector.IsPointInside2D(p) {
			continue
		}
		return sector
	}
	return nil
}
func (b *Body) Construct(data map[string]any) {
	b.Attached.Construct(data)

	b.Pos.Set(concepts.Vector3{0, 0, 0})
	b.Vel.Set(concepts.Vector3{0, 0, 0})
	b.Size.Set(concepts.Vector2{10, 10})
	b.CollisionResponse = Slide
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
	if v, ok := data["CollisionResponse"]; ok {
		c, err := CollisionResponseString(v.(string))
		if err == nil {
			b.CollisionResponse = c
		} else {
			panic(err)
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
	result["MountHeight"] = b.MountHeight
	result["CollisionResponse"] = b.CollisionResponse.String()
	if b.Shadow != BodyShadowNone {
		result["Shadow"] = b.Shadow.String()
	}
	return result
}
