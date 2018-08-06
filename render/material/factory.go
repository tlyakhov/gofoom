package material

import (
	"fmt"
	"reflect"

	"github.com/tlyakhov/gofoom/mapping/material"
	"github.com/tlyakhov/gofoom/render/state"
)

func For(concrete interface{}, s *state.Slice) state.Sampleable {
	if concrete == nil {
		return nil
	}
	switch target := concrete.(type) {
	case *material.LitSampled:
		return NewLitSampledService(target, s)
	case *material.Lit:
		return NewLitService(target, s)
	case *material.Sampled:
		return NewSampledService(target, s)
	case *material.Sky:
		return NewSkyService(target, s)
	case *material.PainfulLitSampled:
		return NewLitSampledService(&target.LitSampled, s)
	default:
		panic(fmt.Sprintf("Tried to get a material service for %v and didn't find one.", reflect.TypeOf(concrete)))
	}
}
