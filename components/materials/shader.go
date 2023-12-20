package materials

import "tlyakhov/gofoom/concepts"

type ShaderStage struct {
	DB        *concepts.EntityComponentDB
	Texture   *concepts.EntityRef `editable:"Texture" edit_type:"Material"`
	Transform concepts.Matrix2    `editable:"Transform"`
	// TODO: implement
	Blend any
}
type Shader struct {
	concepts.Attached `editable:"^"`

	Stages []ShaderStage `editable:"Stages"`
}

var ShaderComponentIndex int

func init() {
	ShaderComponentIndex = concepts.DbTypes().Register(Shader{}, ShaderFromDb)
}

func ShaderFromDb(entity *concepts.EntityRef) *Shader {
	if asserted, ok := entity.Component(ShaderComponentIndex).(*Shader); ok {
		return asserted
	}
	return nil
}

func (s *Shader) Construct(data map[string]any) {
	s.Attached.Construct(data)

	s.Stages = make([]ShaderStage, 0)

	if data == nil {
		return
	}

	if v, ok := data["Stages"]; ok {
		s.Stages = constructShaderStages(s.DB, v)
	}
}

func (s *Shader) Serialize() map[string]any {
	result := s.Attached.Serialize()

	result["Stages"] = serializeShaderStages(s.Stages)
	return result
}

func constructShaderStages(db *concepts.EntityComponentDB, data any) []ShaderStage {
	var result []ShaderStage

	if stages, ok := data.([]any); ok {
		result = make([]ShaderStage, len(stages))
		for i, child := range stages {
			result[i].Construct(db, child.(map[string]any))
		}
	}
	return result
}

func serializeShaderStages(stages []ShaderStage) []map[string]any {
	result := make([]map[string]any, len(stages))
	for i, stage := range stages {
		result[i] = stage.Serialize()
	}
	return result
}

func (s *ShaderStage) Construct(db *concepts.EntityComponentDB, data map[string]any) {
	s.DB = db
	s.Transform = concepts.IdentityMatrix2

	if data == nil {
		return
	}

	if v, ok := data["Texture"]; ok {
		s.Texture = s.DB.DeserializeEntityRef(v)
	}

	if v, ok := data["Transform"]; ok {
		s.Transform.Deserialize(v.([]any))
	}
}

func (s *ShaderStage) Serialize() map[string]any {
	result := make(map[string]any)

	if !s.Texture.Nil() {
		result["Texture"] = s.Texture.Serialize()
	}
	result["Transform"] = s.Transform.Serialize()
	return result
}
