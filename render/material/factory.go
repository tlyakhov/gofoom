package material

import (
	"fmt"
	"reflect"

	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/materials"
	"tlyakhov/gofoom/render/state"
)

func For(concrete interface{}, s *state.Slice) state.Sampleable {
	if concrete == nil {
		return nil
	}
	var mat state.Sampleable
	id := concrete.(concepts.ISerializable).GetBase().ID
	if mat, ok := s.Config.MaterialServiceCache.Load(id); ok {
		return mat.(state.Sampleable)
	}
	switch target := concrete.(type) {
	case *materials.LitSampled:
		mat = NewLitSampledService(target, s)
	case *materials.Lit:
		mat = NewLitService(target, s)
	case *materials.Sampled:
		mat = NewSampledService(target, s)
	case *materials.Sky:
		mat = NewSkyService(target, s)
	case *materials.PainfulLitSampled:
		mat = NewLitSampledService(&target.LitSampled, s)
	default:
		panic(fmt.Sprintf("Tried to get a material service for %v and didn't find one.", reflect.TypeOf(concrete)))
	}
	s.Config.MaterialServiceCache.Store(id, mat)
	return mat
}
