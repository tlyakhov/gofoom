package logic

import (
	"reflect"

	"github.com/tlyakhov/gofoom/mapping"

	"github.com/tlyakhov/gofoom/mapping/material"
)

var (
	TypeMap = map[reflect.Type]reflect.Type{
		reflect.TypeOf(&material.Painful{}): reflect.TypeOf(&Painful{}),
		reflect.TypeOf(&mapping.Sector{}):   reflect.TypeOf(&Sector{}),
		reflect.TypeOf(&mapping.Player{}):   reflect.TypeOf(&Player{}),
	}
)
