package state

import (
	"reflect"
	"strings"
)

type PropertyGridField struct {
	Name             string
	Values           []reflect.Value
	Unique           map[string]reflect.Value
	Type             reflect.Type
	ParentName       string
	Depth            int
	Source           *reflect.StructField
	ParentCollection *reflect.Value
	Parent           any
}

func (f *PropertyGridField) Short() string {
	split := strings.Split(f.Name, "[")
	if len(split) > 1 {
		return "[" + split[len(split)-1]
	}
	return f.Name
}
