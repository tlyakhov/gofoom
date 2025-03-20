// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package state

import (
	"reflect"
	"tlyakhov/gofoom/ecs"
)

type PropertyGridFieldValue struct {
	Entity           ecs.Entity
	Component        ecs.Attachable
	Value            reflect.Value
	ParentCollection *reflect.Value
	Ancestors        []any
}

func (v *PropertyGridFieldValue) Deref() reflect.Value {
	return v.Value.Elem()
}
func (v *PropertyGridFieldValue) Interface() any {
	return v.Value.Elem().Interface()
}

func (v *PropertyGridFieldValue) Parent() any {
	if len(v.Ancestors) == 0 {
		return nil
	}
	return v.Ancestors[len(v.Ancestors)-1]
}
