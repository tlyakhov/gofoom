package materials

import "tlyakhov/gofoom/concepts"

type ShaderStage struct {
	*Shader
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
		s.Stages = s.constructShaderStages(v)
	}
}

func (s *Shader) Serialize() map[string]any {
	result := s.Attached.Serialize()

	result["Stages"] = s.serializeShaderStages()
	return result
}

func (s *Shader) constructShaderStages(data any) []ShaderStage {
	var result []ShaderStage

	if stages, ok := data.([]any); ok {
		result = make([]ShaderStage, len(stages))
		for i, child := range stages {
			result[i].Construct(s, child.(map[string]any))
		}
	}
	return result
}

func (s *Shader) serializeShaderStages() []map[string]any {
	result := make([]map[string]any, len(s.Stages))
	for i, stage := range s.Stages {
		result[i] = stage.Serialize()
	}
	return result
}

func (s *ShaderStage) Construct(shader *Shader, data map[string]any) {
	s.Shader = shader
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
