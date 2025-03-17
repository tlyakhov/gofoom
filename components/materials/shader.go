// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package materials

import (
	"tlyakhov/gofoom/ecs"
)

type Shader struct {
	ecs.Attached `editable:"^"`

	Stages []*ShaderStage `editable:"Stages"`
}

var ShaderCID ecs.ComponentID

func init() {
	ShaderCID = ecs.RegisterComponent(&ecs.Column[Shader, *Shader]{Getter: GetShader})
}

func GetShader(u *ecs.Universe, e ecs.Entity) *Shader {
	if asserted, ok := u.Component(e, ShaderCID).(*Shader); ok {
		return asserted
	}
	return nil
}

func (s *Shader) MultiAttachable() bool { return true }

func (s *Shader) Construct(data map[string]any) {
	s.Attached.Construct(data)

	s.Stages = make([]*ShaderStage, 0)

	if data == nil {
		return
	}

	if v, ok := data["Stages"]; ok {
		s.Stages = ecs.ConstructSlice[*ShaderStage](s.Universe, v, nil)
	}
}

func (s *Shader) Serialize() map[string]any {
	result := s.Attached.Serialize()

	result["Stages"] = ecs.SerializeSlice(s.Stages)
	return result
}
