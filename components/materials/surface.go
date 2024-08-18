// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package materials

import (
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/ecs"
)

/*
	regex:

["](Floor|Ceil|Mid|Lo|Hi)Material["][:] (["]\d+["])
"$1Surface": { "Material": $2 }

["](Mid|Lo|Hi)Surface["]
"$1"
*/

type Surface struct {
	DB          *ecs.ECS
	Material    ecs.Entity                         `editable:"Material" edit_type:"Material"`
	ExtraStages []*ShaderStage                     `editable:"Extra Shader Stages"`
	Transform   ecs.DynamicValue[concepts.Matrix2] `editable:"Transform"`
}

func (s *Surface) Construct(db *ecs.ECS, data map[string]any) {
	s.DB = db
	s.ExtraStages = make([]*ShaderStage, 0)
	s.Transform.Construct(nil)

	if data == nil {
		return
	}

	if v, ok := data["ExtraStages"]; ok {
		s.ExtraStages = ecs.ConstructSlice[*ShaderStage](db, v)
	}
	if v, ok := data["Material"]; ok {
		s.Material, _ = ecs.ParseEntity(v.(string))
	}
	if v, ok := data["Transform"]; ok {
		if v2, ok2 := v.([]any); ok2 {
			v = map[string]any{"Original": v2}
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
		result["Material"] = s.Material.Format()
	}

	if !s.Transform.Original.IsIdentity() {
		result["Transform"] = s.Transform.Serialize()
	}

	return result
}
