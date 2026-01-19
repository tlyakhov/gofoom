// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"reflect"
	"strings"
)

// ComponentFlags represents flags that can be associated with a component.
//
//go:generate go run github.com/dmarkham/enumer -type=ComponentFlags -json
type ComponentFlags uint16

const (
	// ComponentActive indicates that the component should be processed by controllers.
	ComponentActive ComponentFlags = 1 << iota
	// ComponentNoSave indicates that the component should not be saved to disk.
	ComponentNoSave
	// ComponentHideInEditor indicates that the component should be hidden in
	// the editor.
	ComponentHideInEditor
	// ComponentHideEntityInEditor indicates that the entire entity this
	// component is attached to should be hidden in the editor.
	ComponentHideEntityInEditor
	// ComponentLockedInEditor indicates that the component should be locked in
	// the editor, preventing modifications.
	ComponentLockedInEditor
	// ComponentLockedEntityInEditor indicates that the whole entity should be
	// locked in the editor, preventing modifications.
	ComponentLockedEntityInEditor
)

// ComponentInternal is a combination of flags indicating that the component is
// internal to the engine and should not be saved or modified by the user.
const ComponentInternal = ComponentNoSave | ComponentHideInEditor | ComponentLockedInEditor

// EntityInternal is a combination of flags indicating that the entire entity is
// internal to the engine and should not be saved or modified by the user.
const EntityInternal = ComponentInternal | ComponentHideEntityInEditor | ComponentLockedEntityInEditor

//go:generate go run github.com/dmarkham/enumer -type=FieldFlags -json
type FieldFlags uint16

const (
	FieldNonCacheable FieldFlags = 1 << iota
	FieldNonTraversable
	FieldShallowCacheable
)

func FieldFlagsFromTag(tag reflect.StructTag) FieldFlags {
	root := tag.Get("ecs")
	if root == "" {
		return 0
	}
	var result FieldFlags
	for flag := range strings.SplitSeq(root, ",") {
		switch flag {
		case "non-cacheable":
			result |= FieldNonCacheable
		case "non-traversable":
			result |= FieldNonTraversable
		case "shallow-cacheable":
			result |= FieldShallowCacheable
		}
	}
	return result
}
