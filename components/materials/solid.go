// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package materials

import (
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/dynamic"
	"tlyakhov/gofoom/ecs"
)

type Solid struct {
	ecs.Attached `editable:"^"`
	Diffuse      dynamic.DynamicValue[concepts.Vector4] `editable:"Color"`
}

var SolidCID ecs.ComponentID

func init() {
	SolidCID = ecs.RegisterComponent(&ecs.Column[Solid, *Solid]{Getter: GetSolid})
}

func (x *Solid) ComponentID() ecs.ComponentID {
	return SolidCID
}
func GetSolid(u *ecs.Universe, e ecs.Entity) *Solid {
	if asserted, ok := u.Component(e, SolidCID).(*Solid); ok {
		return asserted
	}
	return nil
}

func (s *Solid) MultiAttachable() bool { return true }

func (s *Solid) OnDelete() {
	defer s.Attached.OnDelete()
	if s.Universe != nil {
		s.Diffuse.Detach(s.Universe.Simulation)
	}
}
func (s *Solid) OnAttach(u *ecs.Universe) {
	s.Attached.OnAttach(u)
	s.Diffuse.Attach(s.Universe.Simulation)
}

func (s *Solid) Construct(data map[string]any) {
	s.Attached.Construct(data)
	s.Diffuse.Construct(nil)

	if data == nil {
		return
	}

	if v, ok := data["Diffuse"]; ok {
		s.Diffuse.Construct(v.(map[string]any))
	}
}

func (s *Solid) Serialize() map[string]any {
	result := s.Attached.Serialize()
	result["Diffuse"] = s.Diffuse.Serialize()
	return result
}
