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

func (s *Shader) Shareable() bool { return true }

func (s *Shader) Construct(data map[string]any) {
	s.Attached.Construct(data)

	s.Stages = make([]*ShaderStage, 0)

	if data == nil {
		return
	}

	if v, ok := data["Stages"]; ok {
		s.Stages = ecs.ConstructSlice[*ShaderStage](v, nil)
	}
}

func (s *Shader) Serialize() map[string]any {
	result := s.Attached.Serialize()

	result["Stages"] = ecs.SerializeSlice(s.Stages)
	return result
}
