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

var ShaderComponentIndex int

func init() {
	ShaderComponentIndex = ecs.Types().Register(Shader{}, ShaderFromDb)
}

func ShaderFromDb(db *ecs.ECS, e ecs.Entity) *Shader {
	if asserted, ok := db.Component(e, ShaderComponentIndex).(*Shader); ok {
		return asserted
	}
	return nil
}

func (s *Shader) Construct(data map[string]any) {
	s.Attached.Construct(data)

	s.Stages = make([]*ShaderStage, 0)

	if data == nil {
		return
	}

	if v, ok := data["Stages"]; ok {
		s.Stages = ecs.ConstructSlice[*ShaderStage](s.DB, v)
	}
}

func (s *Shader) Serialize() map[string]any {
	result := s.Attached.Serialize()

	result["Stages"] = ecs.SerializeSlice(s.Stages)
	return result
}
