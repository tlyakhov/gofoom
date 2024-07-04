// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package materials

import (
	"tlyakhov/gofoom/concepts"
)

/*
	regex:

["](Floor|Ceil|Mid|Lo|Hi)Material["][:] (["]\d+["])
"$1Surface": { "Material": $2 }

["](Mid|Lo|Hi)Surface["]
"$1"
*/
type SurfaceStretch int

//go:generate go run github.com/dmarkham/enumer -type=SurfaceStretch -json
const (
	StretchNone SurfaceStretch = iota
	StretchAspect
	StretchFill
)

type Surface struct {
	DB       *concepts.EntityComponentDB
	Material concepts.Entity `editable:"Material" edit_type:"Material"`
	// Do we need both this AND the transform? not sure
	Stretch     SurfaceStretch   `editable:"Stretch"`
	ExtraStages []*ShaderStage   `editable:"Extra Shader Stages"`
	Transform   concepts.Matrix2 `editable:"Transform"`
}

func (s *Surface) Construct(db *concepts.EntityComponentDB, data map[string]any) {
	s.DB = db
	s.ExtraStages = make([]*ShaderStage, 0)
	s.Transform = concepts.IdentityMatrix2

	if data == nil {
		return
	}

	if v, ok := data["ExtraStages"]; ok {
		s.ExtraStages = concepts.ConstructSlice[*ShaderStage](db, v)
	}
	if v, ok := data["Material"]; ok {
		s.Material, _ = concepts.DeserializeEntity(v.(string))
	}
	if v, ok := data["Transform"]; ok {
		s.Transform.Deserialize(v.([]any))
	}
	if v, ok := data["Stretch"]; ok {
		ms, err := SurfaceStretchString(v.(string))
		if err == nil {
			s.Stretch = ms
		} else {
			panic(err)
		}
	}
}

func (s *Surface) Serialize() map[string]any {
	result := make(map[string]any)

	if len(s.ExtraStages) > 0 {
		result["ExtraStages"] = concepts.SerializeSlice(s.ExtraStages)
	}

	if s.Material != 0 {
		result["Material"] = s.Material.Serialize()
	}

	if s.Stretch != StretchNone {
		result["Stretch"] = s.Stretch.String()
	}

	if !s.Transform.IsIdentity() {
		result["Transform"] = s.Transform.Serialize()
	}

	return result
}
