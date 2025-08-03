// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package materials

import (
	"fmt"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/dynamic"
	"tlyakhov/gofoom/ecs"
)

type Surface struct {
	Material    ecs.Entity                             `editable:"Material" edit_type:"Material"`
	ExtraStages []*ShaderStage                         `editable:"Extra Shader Stages"`
	Transform   dynamic.DynamicValue[concepts.Matrix2] `editable:"ℝ²→ℝ²"`
}

func (s *Surface) String() string {
	if len(s.ExtraStages) == 0 {
		return fmt.Sprintf("Surface (%v)", s.Material)
	}
	return fmt.Sprintf("Surface (%v) + %v extra stages", s.Material, len(s.ExtraStages))
}

func (s *Surface) Construct(data map[string]any) {
	s.ExtraStages = make([]*ShaderStage, 0)
	s.Transform.Construct(nil)

	if data == nil {
		return
	}

	if v, ok := data["ExtraStages"]; ok {
		s.ExtraStages = ecs.ConstructSlice[*ShaderStage](v, nil)
	}
	if v, ok := data["Material"]; ok {
		s.Material, _ = ecs.ParseEntity(v.(string))
	}
	if v, ok := data["Transform"]; ok {
		if v2, ok2 := v.([]any); ok2 {
			v = map[string]any{"Spawn": v2}
		}
		s.Transform.Construct(v.(map[string]any))
	}
}

func (s *Surface) Serialize() map[string]any {
	result := make(map[string]any)

	if len(s.ExtraStages) > 0 {
		result["ExtraStages"] = ecs.SerializeSlice(s.ExtraStages)
	}

	if s.Material != 0 {
		result["Material"] = s.Material.Serialize()
	}

	if !s.Transform.Spawn.IsIdentity() {
		result["Transform"] = s.Transform.Serialize()
	}

	return result
}
