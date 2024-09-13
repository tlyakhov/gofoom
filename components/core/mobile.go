// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package core

import (
	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/dynamic"
	"tlyakhov/gofoom/ecs"

	"tlyakhov/gofoom/concepts"
)

// TODO: Separate out "Collidable" component?
type Mobile struct {
	ecs.Attached `editable:"^"`
	Vel          dynamic.DynamicValue[concepts.Vector3]
	Force        concepts.Vector3
	Mass         float64 `editable:"Mass"`

	// "Bounciness" (0 = inelastic, 1 = perfectly elastic)
	Elasticity  float64           `editable:"Elasticity"`
	CrBody      CollisionResponse `editable:"Collision (Body)"`
	CrPlayer    CollisionResponse `editable:"Collision (Player)"`
	CrWall      CollisionResponse `editable:"Collision (Wall)"`
	MountHeight float64           `editable:"Mount Height"`
}

var MobileCID ecs.ComponentID

func init() {
	MobileCID = ecs.RegisterComponent(&ecs.Column[Mobile, *Mobile]{Getter: GetMobile}, "")
}

func GetMobile(db *ecs.ECS, e ecs.Entity) *Mobile {
	if asserted, ok := db.Component(e, MobileCID).(*Mobile); ok {
		return asserted
	}
	return nil
}

func (m *Mobile) String() string {
	return "Mobile: " + m.Vel.Now.StringHuman()
}

func (m *Mobile) OnDetach() {
	if m.ECS != nil {
		m.Vel.Detach(m.ECS.Simulation)
	}
	m.Attached.OnDetach()
}

func (m *Mobile) AttachECS(db *ecs.ECS) {
	if m.ECS != db {
		m.OnDetach()
	}
	m.Attached.AttachECS(db)
	m.Vel.Attach(db.Simulation)
}

func (m *Mobile) Construct(data map[string]any) {
	m.Attached.Construct(data)

	m.Vel.Construct(nil)

	m.Elasticity = 0.5
	m.CrBody = CollideNone
	m.CrPlayer = CollideNone
	m.CrWall = CollideSeparate
	m.MountHeight = constants.PlayerMountHeight

	if data == nil {
		return
	}

	if v, ok := data["Vel"]; ok {
		v3 := v.(map[string]any)
		if _, ok2 := v3["X"]; ok2 {
			v3 = map[string]any{"Original": v3}
		}
		m.Vel.Construct(v3)
	}

	if v, ok := data["MountHeight"]; ok {
		m.MountHeight = v.(float64)
	}
	if v, ok := data["Mass"]; ok {
		m.Mass = v.(float64)
	}
	if v, ok := data["Elasticity"]; ok {
		m.Elasticity = v.(float64)
	}
	if v, ok := data["CrMoving"]; ok {
		c, err := CollisionResponseString(v.(string))
		if err == nil {
			m.CrBody = c
		}
	}
	if v, ok := data["CrPlayer"]; ok {
		c, err := CollisionResponseString(v.(string))
		if err == nil {
			m.CrPlayer = c
		}
	}
	if v, ok := data["CrWall"]; ok {
		c, err := CollisionResponseString(v.(string))
		if err == nil {
			m.CrWall = c
		}
	}
}

func (m *Mobile) Serialize() map[string]any {
	result := m.Attached.Serialize()
	result["Vel"] = m.Vel.Serialize()
	result["Mass"] = m.Mass
	result["Elasticity"] = m.Elasticity
	result["MountHeight"] = m.MountHeight
	result["CrMoving"] = m.CrBody.String()
	result["CrPlayer"] = m.CrPlayer.String()
	result["CrWall"] = m.CrWall.String()
	return result
}
