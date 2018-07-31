package render

import (
	"reflect"

	"github.com/tlyakhov/gofoom/mapping/material"
)

var (
	typeMap = map[reflect.Type]reflect.Type{
		reflect.TypeOf(&material.Lit{}):        reflect.TypeOf(&Lit{}),
		reflect.TypeOf(&material.Sampled{}):    reflect.TypeOf(&Sampled{}),
		reflect.TypeOf(&material.LitSampled{}): reflect.TypeOf(&LitSampled{}),
		reflect.TypeOf(&material.Sky{}):        reflect.TypeOf(&Sky{}),
	}
)
