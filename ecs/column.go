// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"log"
	"reflect"
)

const chunkSize = 64

// Note, we can't use a slice for the data, because we use long-lived pointers to
// components, and slices can reallocate the backing array when resized.
// See https://utcc.utoronto.ca/~cks/space/blog/programming/GoSlicesVsPointers
type componentChunk[T any, PT GenericAttachable[T]] [chunkSize]T

type Column[T any, PT GenericAttachable[T]] struct {
	*ECS
	data  []*componentChunk[T, PT]
	Index int
	// TODO: Replace with bitmap.Bitmap fill list?
	Length int
	Getter func(ecs *ECS, e Entity) PT

	typeOfT reflect.Type
}

// No bounds checking for performance. This should be inlined
func (col *Column[T, PT]) Value(index int) PT {
	return &col.data[index/chunkSize][index%chunkSize]
}

func (c *Column[T, PT]) Attachable(index int) Attachable {
	return c.Value(index)
}

func (col *Column[T, PT]) Detach(index int) {
	if index < col.Length {
		if index != col.Length-1 {
			// swap with last element
			*col.Value(index) = *col.Value(col.Length - 1)
			col.Value(index).SetColumnIndex(index)
		}
		col.Length--
		// If we've detached the last in a chunk, remove it
		if len(col.data) > col.Length/chunkSize {
			col.data = col.data[:len(col.data)-1]
		}
	} else {
		log.Printf("ecs.Column.Detach: found component index %v, but component list is too short.", index)
	}
}

func (col *Column[T, PT]) Add(component Attachable) Attachable {
	chunk := col.Length / chunkSize
	if chunk >= len(col.data) {
		col.data = append(col.data, new(componentChunk[T, PT]))
	}
	index := col.Length % chunkSize

	if component != nil {
		col.data[chunk][index] = *(component.(PT))
	}

	component = PT(&col.data[chunk][index])
	component.SetColumnIndex(index)
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

func (c *Column[T, PT]) String() string {
	return c.typeOfT.String()
}
