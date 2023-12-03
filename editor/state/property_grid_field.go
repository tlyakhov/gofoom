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
	EditType         string
	ParentName       string
	Depth            int
	Source           *reflect.StructField
	ParentCollection *reflect.Value
	Parent           any
}

func (f *PropertyGridField) Short() string {
	result := f.Name
	reduced := false
	for len(result) > 60 {
		reduced = true
		split := strings.Split(result, ".")
		if len(split) == 1 {
			break
		}
		result = strings.Join(split[1:], ".")
	}
	if reduced {
		result = "{...}." + result
	}
	return result
	/*split := strings.Split(f.Name, "[")
	if len(split) > 1 {
		return "[" + split[len(split)-1]
	}
	return f.Name*/
}
