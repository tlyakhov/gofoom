// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package state

import (
	"reflect"
	"strings"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/ecs"
)

var EmbeddedTypes = [...]string{
	reflect.TypeFor[*ecs.DynamicValue[float64]]().String(),
	reflect.TypeFor[*ecs.DynamicValue[int]]().String(),
	reflect.TypeFor[*ecs.DynamicValue[concepts.Vector2]]().String(),
	reflect.TypeFor[*ecs.DynamicValue[concepts.Vector3]]().String(),
	reflect.TypeFor[*ecs.DynamicValue[concepts.Vector4]]().String(),
	reflect.TypeFor[*ecs.DynamicValue[concepts.Matrix2]]().String(),
	reflect.TypeFor[*core.Script]().String(),
	reflect.TypeFor[*materials.Surface]().String(),
	reflect.TypeFor[*materials.ShaderStage]().String(),
	reflect.TypeFor[*materials.Sprite]().String(),
	reflect.TypeFor[**ecs.Animation[float64]]().String(),
	reflect.TypeFor[**ecs.Animation[int]]().String(),
	reflect.TypeFor[**ecs.Animation[concepts.Vector2]]().String(),
	reflect.TypeFor[**ecs.Animation[concepts.Vector3]]().String(),
	reflect.TypeFor[**ecs.Animation[concepts.Vector4]]().String(),
	reflect.TypeFor[**ecs.Animation[concepts.Matrix2]]().String(),
}

type PropertyGridField struct {
	Name       string
	ParentName string
	Type       reflect.Type
	EditType   string
	Source     *reflect.StructField
	Sort       int
	Depth      int
	Parent     *PropertyGridField

	Values []*PropertyGridFieldValue
	Unique map[string]reflect.Value
}

func (f *PropertyGridField) Short() string {
	result := f.Name
	reduced := false
	for len(result) > 32 {
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
