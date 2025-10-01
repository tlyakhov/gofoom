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

func (s *Solid) MultiAttachable() bool { return true }

func (s *Solid) OnDelete() {
	defer s.Attached.OnDelete()
	if s.IsAttached() {
		s.Diffuse.Detach(ecs.Simulation)
	}
}
func (s *Solid) OnAttach() {
	s.Attached.OnAttach()
	s.Diffuse.Attach(ecs.Simulation)
}

func (s *Solid) Construct(data map[string]any) {
	s.Attached.Construct(data)
	s.Diffuse.Construct(nil)

	if data == nil {
		return
	}

	if v, ok := data["Diffuse"]; ok {
		s.Diffuse.Construct(v)
	}
}

func (s *Solid) Serialize() map[string]any {
	result := s.Attached.Serialize()
	result["Diffuse"] = s.Diffuse.Serialize()
	return result
}
