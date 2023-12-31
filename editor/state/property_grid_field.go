package state

import (
	"reflect"
	"strings"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"
)

var EmbeddedTypes = [...]string{
	reflect.TypeOf(&concepts.SimVariable[concepts.Vector2]{}).String(),
	reflect.TypeOf(&concepts.SimVariable[concepts.Vector3]{}).String(),
	reflect.TypeOf(&concepts.SimVariable[concepts.Vector4]{}).String(),
	reflect.TypeOf(&concepts.SimVariable[float64]{}).String(),
	reflect.TypeOf(&core.Script{}).String(),
	reflect.TypeOf(&materials.Surface{}).String(),
	reflect.TypeOf(&materials.ShaderStage{}).String(),
}

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
	for len(result) > 40 {
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

func (f *PropertyGridField) IsEmbeddedType() bool {
	for _, t := range EmbeddedTypes {
		//log.Printf("%v - %v", f.Short(), f.Type.String())
		if f.Type.String() == t {
			return true
		}
	}
	return false
}
