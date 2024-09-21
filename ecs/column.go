// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"log"
	"reflect"

	"github.com/kelindar/bitmap"
)

const chunkSize = 64

// Note, we can't use a slice for the data, because we use long-lived pointers to
// components, and slices can reallocate the backing array when resized.
// See https://utcc.utoronto.ca/~cks/space/blog/programming/GoSlicesVsPointers
type componentChunk[T any, PT GenericAttachable[T]] [chunkSize]T

type Column[T any, PT GenericAttachable[T]] struct {
	*ECS
	Index  int
	Length int
	Getter func(ecs *ECS, e Entity) PT

	data        []*componentChunk[T, PT]
	fill        bitmap.Bitmap
	typeOfT     reflect.Type
	componentID ComponentID
}

func (col *Column[T, PT]) From(source AttachableColumn, ecs *ECS) {
	placeholder := source.(*Column[T, PT])
	col.ECS = ecs
	col.typeOfT = placeholder.typeOfT
	col.componentID = placeholder.componentID
	col.Getter = placeholder.Getter
}

// No bounds checking for performance. This should always be inlined
func (col *Column[T, PT]) Value(index int) PT {
	ptr := PT(&(col.data[index/chunkSize][index%chunkSize]))
	if ptr.GetECS() == nil {
		return nil
	}
	return ptr
}

// No bounds checking for performance. This should always be inlined
// Duplicates code in .Value() because the return type is different here
func (col *Column[T, PT]) Attachable(index int) Attachable {
	ptr := PT(&(col.data[index/chunkSize][index%chunkSize]))
	if ptr.GetECS() == nil {
		return nil
	}
	return ptr
}

func (col *Column[T, PT]) Detach(index int) {
	if index >= col.Cap() {
		log.Printf("ecs.Column.Detach: found component index %v, but component list is too short.", index)
		return
	}
	// It's tempting to copy the last element into the detached one and
	// shrink the chunk, but unfortunately, we can't do that because it
	// would affect any pointers to that last element (which would now
	// become invalid). Instead we use a bitmap fill list.
	col.fill.Remove(uint32(index))
	col.Length--
	// TODO: Remove empty chunks
}

func (col *Column[T, PT]) Add(component Attachable) Attachable {
	var nextFree uint32
	var found bool
	// First try an empty slot in the fill list
	if nextFree, found = col.fill.MinZero(); !found {
		// No empty slots, put it at the end
		nextFree = uint32(col.Cap())
	}

	chunk := nextFree / chunkSize
	if chunk >= uint32(len(col.data)) {
		// Create a new chunk
		col.data = append(col.data, new(componentChunk[T, PT]))
	}

	indexInChunk := nextFree % chunkSize

	if component != nil {
		col.data[chunk][indexInChunk] = *(component.(PT))
	}

	col.fill.Set(nextFree)
	component = PT(&col.data[chunk][indexInChunk])
	component.SetColumnIndex(int(nextFree))
	component.AttachECS(col.ECS)
	col.Length++
	return component
}

func (col *Column[T, PT]) Replace(component Attachable, index int) Attachable {
	if component == nil {
		return col.Value(index)
	}
	ptr := col.Value(index)
	*ptr = *(component.(PT))
	component = ptr
	component.SetColumnIndex(index)
	component.AttachECS(col.ECS)
	return component
}

func (c *Column[T, PT]) New() Attachable {
	var x T
	return PT(&x)
}

func (c *Column[T, PT]) Type() reflect.Type {
	return c.typeOfT
}

func (c *Column[T, PT]) Len() int {
	return c.Length
}

func (c *Column[T, PT]) Cap() int {
	return len(c.data) * chunkSize
}

func (c *Column[T, PT]) ID() ComponentID {
	return c.componentID
}

func (c *Column[T, PT]) String() string {
	return c.typeOfT.String()
}
