package materials

import (
	"tlyakhov/gofoom/concepts"
)

/*
	regex:

["](Floor|Ceil|Mid|Lo|Hi)Material["][:] (["]\d+["])
"$1Surface": { "Material": $2 }
*/
type SurfaceScale int

//go:generate go run github.com/dmarkham/enumer -type=SurfaceScale -json
const (
	ScaleNone SurfaceScale = iota
	ScaleHeight
	ScaleWidth
	ScaleAll
)

type Surface struct {
	DB       *concepts.EntityComponentDB
	Material *concepts.EntityRef `editable:"Material" edit_type:"Material"`
	// Do we need both this AND the transform? not sure
	Scale       SurfaceScale     `editable:"Scale"`
	ExtraStages []ShaderStage    `editable:"Extra Shader Stages"`
	Transform   concepts.Matrix2 `editable:"Transform"`
}

func (s *Surface) Construct(db *concepts.EntityComponentDB, data map[string]any) {
	s.DB = db
	s.ExtraStages = make([]ShaderStage, 0)
	s.Transform = concepts.IdentityMatrix2

	if data == nil {
		return
	}

	if v, ok := data["ExtraStages"]; ok {
		s.ExtraStages = constructShaderStages(db, v)
	}
	if v, ok := data["Material"]; ok {
		s.Material = s.DB.DeserializeEntityRef(v)
	}
	if v, ok := data["Scale"]; ok {
		ms, err := SurfaceScaleString(v.(string))
		if err == nil {
			s.Scale = ms
		} else {
			panic(err)
		}
	}
}

func (s *Surface) Serialize() map[string]any {
	result := make(map[string]any)

	if len(s.ExtraStages) > 0 {
		result["ExtraStages"] = serializeShaderStages(s.ExtraStages)
	}

	if !s.Material.Nil() {
		result["Material"] = s.Material.Serialize()
	}

	if s.Scale != ScaleNone {
		result["Scale"] = s.Scale.String()
	}

	return result
}
