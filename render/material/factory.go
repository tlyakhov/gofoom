package material

import (
	"fmt"
	"reflect"

	"tlyakhov/gofoom/materials"
	"tlyakhov/gofoom/render/state"
)

func For(concrete interface{}, s *state.Slice) state.Sampleable {
	if concrete == nil {
		return nil
	}
	switch target := concrete.(type) {
	case *materials.LitSampled:
		return NewLitSampledService(target, s)
	case *materials.Lit:
		return NewLitService(target, s)
	case *materials.Sampled:
		return NewSampledService(target, s)
	case *materials.Sky:
		return NewSkyService(target, s)
	case *materials.PainfulLitSampled:
		return NewLitSampledService(&target.LitSampled, s)
	default:
		panic(fmt.Sprintf("Tried to get a material service for %v and didn't find one.", reflect.TypeOf(concrete)))
	}
}
