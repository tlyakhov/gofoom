// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package state

import (
	"reflect"
	"strings"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"
)

var EmbeddedTypes = [...]string{
	concepts.ReflectType[*concepts.SimVariable[float64]]().String(),
	concepts.ReflectType[*concepts.SimVariable[int]]().String(),
	concepts.ReflectType[*concepts.SimVariable[concepts.Vector2]]().String(),
	concepts.ReflectType[*concepts.SimVariable[concepts.Vector3]]().String(),
	concepts.ReflectType[*concepts.SimVariable[concepts.Vector4]]().String(),
	concepts.ReflectType[*concepts.SimVariable[concepts.Matrix2]]().String(),
	concepts.ReflectType[*core.Script]().String(),
	concepts.ReflectType[*materials.Surface]().String(),
	concepts.ReflectType[*materials.ShaderStage]().String(),
	concepts.ReflectType[*materials.Sprite]().String(),
	concepts.ReflectType[**concepts.Animation[float64]]().String(),
	concepts.ReflectType[**concepts.Animation[int]]().String(),
	concepts.ReflectType[**concepts.Animation[concepts.Vector2]]().String(),
	concepts.ReflectType[**concepts.Animation[concepts.Vector3]]().String(),
	concepts.ReflectType[**concepts.Animation[concepts.Vector4]]().String(),
	concepts.ReflectType[**concepts.Animation[concepts.Matrix2]]().String(),
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
