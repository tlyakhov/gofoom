package materials

import "tlyakhov/gofoom/concepts"

type ShaderStage struct {
	DB        *concepts.EntityComponentDB
	Texture   *concepts.EntityRef `editable:"Texture" edit_type:"Material"`
	Transform concepts.Matrix2    `editable:"Transform"`
	// TODO: implement
	Blend any
}

func (s *ShaderStage) Construct(data map[string]any) {
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

func (s *ShaderStage) SetDB(db *concepts.EntityComponentDB) {
	s.DB = db
}
