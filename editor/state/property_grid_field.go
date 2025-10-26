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
	"tlyakhov/gofoom/ecs"
)

var EmbeddedTypes = map[string]struct{}{
	reflect.TypeFor[*core.Script]().String():      {},
	reflect.TypeFor[*core.SectorPlane]().String(): {},

	reflect.TypeFor[*dynamic.Spawned[float64]]().String():          {},
	reflect.TypeFor[*dynamic.Spawned[int]]().String():              {},
	reflect.TypeFor[*dynamic.Spawned[concepts.Vector2]]().String(): {},
	reflect.TypeFor[*dynamic.Spawned[concepts.Vector3]]().String(): {},
	reflect.TypeFor[*dynamic.Spawned[concepts.Vector4]]().String(): {},
	reflect.TypeFor[*dynamic.Spawned[concepts.Matrix2]]().String(): {},

	reflect.TypeFor[*dynamic.DynamicValue[float64]]().String():          {},
	reflect.TypeFor[*dynamic.DynamicValue[int]]().String():              {},
	reflect.TypeFor[*dynamic.DynamicValue[concepts.Vector2]]().String(): {},
	reflect.TypeFor[*dynamic.DynamicValue[concepts.Vector3]]().String(): {},
	reflect.TypeFor[*dynamic.DynamicValue[concepts.Vector4]]().String(): {},
	reflect.TypeFor[*dynamic.DynamicValue[concepts.Matrix2]]().String(): {},

	reflect.TypeFor[*materials.Surface]().String():     {},
	reflect.TypeFor[*materials.ShaderStage]().String(): {},
	reflect.TypeFor[*materials.Sprite]().String():      {},

	reflect.TypeFor[**dynamic.Animation[float64]]().String():          {},
	reflect.TypeFor[**dynamic.Animation[int]]().String():              {},
	reflect.TypeFor[**dynamic.Animation[concepts.Vector2]]().String(): {},
	reflect.TypeFor[**dynamic.Animation[concepts.Vector3]]().String(): {},
	reflect.TypeFor[**dynamic.Animation[concepts.Vector4]]().String(): {},
	reflect.TypeFor[**dynamic.Animation[concepts.Matrix2]]().String(): {},
}

type PropertyGridField struct {
	Name       string
	ParentName string
	Type       reflect.Type
	EditType   string
	Source     *reflect.StructField
	Sort       int
	Depth      int
	SliceIndex int
	Parent     *PropertyGridField

	Values []*PropertyGridFieldValue
	Unique map[string]reflect.Value
}

func (f *PropertyGridField) Disabled() bool {
	if len(f.Values) == 0 {
		return false
	}

	// TODO: Reconsider this logic... need to be careful
	// If any of the user's selected entities are internal, keep enabled.
	// ...aka if all of the selected entities are external, disable.
	// Also, if all of the selected entities have an ecs.Linked entity with this
	// component, disable.
	externalEntitiesOnly := true
	externalComponentsOnly := true
	linkedComponentsOnly := true
	for _, v := range f.Values {
		externalEntitiesOnly = externalEntitiesOnly && v.Entity.IsExternal()
		externalComponentsOnly = externalComponentsOnly && v.Component.Base().IsExternal()
		if linked := ecs.GetLinked(v.Entity); linked != nil {
			if linked.SourceComponents.Get(v.Component.ComponentID()) == nil {
				linkedComponentsOnly = false
			}
		} else {
			linkedComponentsOnly = false
		}
	}
	return externalEntitiesOnly || externalComponentsOnly || linkedComponentsOnly
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
		result = "{…}." + result
	}
	return result
}

func (f *PropertyGridField) IsEmbeddedType() bool {
	_, ok := EmbeddedTypes[f.Type.String()]
	return ok
}
