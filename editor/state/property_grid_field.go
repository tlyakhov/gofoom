// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package state

import (
	"reflect"
	"strings"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/dynamic"
)

var EmbeddedTypes = [...]string{
	reflect.TypeFor[*core.Script]().String(),
	reflect.TypeFor[*core.SectorPlane]().String(),
	reflect.TypeFor[*dynamic.DynamicValue[float64]]().String(),
	reflect.TypeFor[*dynamic.DynamicValue[int]]().String(),
	reflect.TypeFor[*dynamic.DynamicValue[concepts.Vector2]]().String(),
	reflect.TypeFor[*dynamic.DynamicValue[concepts.Vector3]]().String(),
	reflect.TypeFor[*dynamic.DynamicValue[concepts.Vector4]]().String(),
	reflect.TypeFor[*dynamic.DynamicValue[concepts.Matrix2]]().String(),
	reflect.TypeFor[*materials.Surface]().String(),
	reflect.TypeFor[*materials.ShaderStage]().String(),
	reflect.TypeFor[*materials.Sprite]().String(),
	reflect.TypeFor[**dynamic.Animation[float64]]().String(),
	reflect.TypeFor[**dynamic.Animation[int]]().String(),
	reflect.TypeFor[**dynamic.Animation[concepts.Vector2]]().String(),
	reflect.TypeFor[**dynamic.Animation[concepts.Vector3]]().String(),
	reflect.TypeFor[**dynamic.Animation[concepts.Vector4]]().String(),
	reflect.TypeFor[**dynamic.Animation[concepts.Matrix2]]().String(),
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
	s := f.Type.String()
	for _, t := range EmbeddedTypes {
		//log.Printf("%v - %v", f.Short(), f.Type.String())
		if s == t {
			return true
		}
	}
	return false
}
