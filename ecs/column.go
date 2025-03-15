// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"log"
	"reflect"

	"github.com/kelindar/bitmap"
)

// chunkSize is the size of each chunk in the column. Columns are split into chunks to improve memory management and reduce the cost of resizing.
const chunkSize = 64

// componentChunk represents a fixed-size array of components.
// Note: we can't use a slice for the data, because we use long-lived pointers to
// components, and slices can reallocate the backing array when resized, invalidating existing pointers.
// See https://utcc.utoronto.ca/~cks/space/blog/programming/GoSlicesVsPointers
type componentChunk[T any, PT GenericAttachable[T]] [chunkSize]T

// Column is a sparse columnar storage for components of a specific type.
// It stores components in chunks of fixed size, allowing for efficient access and iteration.
// The type parameters are:
//   - T: The type of the component data.
//   - PT: A pointer to the component type, which must implement the GenericAttachable interface.
type Column[T any, PT GenericAttachable[T]] struct {
	// ECS is a pointer to the ECS instance that manages this column.
	*ECS
	// Index is the index of this column within the ECS.
	Index int
	// Length is the number of components currently stored in the column.
	Length int
	// Getter is a function that retrieves a component of this type for a given entity.
	Getter func(ecs *ECS, e Entity) PT

	// data is a slice of pointers to component chunks, where the actual component data is stored.
	data []*componentChunk[T, PT]
	// fill is a bitmap that tracks which slots in the column are occupied by components.
	fill bitmap.Bitmap
	// typeOfT is the reflect.Type of the component data.
	typeOfT reflect.Type
	// componentID is the unique identifier for the component type this column stores.
	componentID ComponentID
}

// From initializes a column from another column of the same type, copying metadata and setting the ECS instance.
func (col *Column[T, PT]) From(source AttachableColumn, ecs *ECS) {
	placeholder := source.(*Column[T, PT])
	col.ECS = ecs
	col.typeOfT = placeholder.typeOfT
	col.componentID = placeholder.componentID
	col.Getter = placeholder.Getter
}

// Value retrieves the component at the given index in the column.
// It performs no bounds checking for performance reasons.
func (col *Column[T, PT]) Value(index int) PT {
	// No bounds checking for performance. This should always be inlined
	ptr := PT(&(col.data[index/chunkSize][index%chunkSize]))
	if ptr.GetECS() == nil {
		return nil
	}
	return ptr
}

// Attachable retrieves the component at the given index as an Attachable interface.
// It performs no bounds checking for performance reasons.
func (col *Column[T, PT]) Attachable(index int) Attachable {
	// No bounds checking for performance. This should always be inlined
	// Duplicates code in .Value() because the return type is different here and nil
	// in golang behaves idiosyncratically
	ptr := PT(&(col.data[index/chunkSize][index%chunkSize]))
	if ptr.GetECS() == nil {
		return nil
	}
	return ptr
}

// Detach removes the component at the given index from the column.
// It uses a fill bitmap to mark the slot as empty instead of shifting elements.
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

// AddTyped adds a component to the column, automatically handling the Attachable interface conversion.
func (col *Column[T, PT]) AddTyped(component *PT) {
	attachable := Attachable(*component)
	col.Add(&attachable)
	*component = attachable.(PT)
}

// Add adds a component to the column at the next available slot.
// It uses a fill bitmap to find the next free index.
func (col *Column[T, PT]) Add(component *Attachable) {
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

	if *component != nil {
		col.data[chunk][indexInChunk] = *(*component).(PT)
	}

	col.fill.Set(nextFree)
	*component = PT(&col.data[chunk][indexInChunk])
	(*component).Base().indexInColumn = (int(nextFree))
	col.Length++
}

// Replace replaces the component at the given index with the provided component.
func (col *Column[T, PT]) Replace(component *Attachable, index int) {
	if *component == nil {
		*component = col.Value(index)
		return
	}
	ptr := col.Value(index)
	*ptr = *((*component).(PT))
	*component = ptr
	ptr.Base().indexInColumn = index
}

// New creates a new Attachable component of the type stored in this column.
func (c *Column[T, PT]) New() Attachable {
	var component T
	attachable := PT(&component)
	attachable.Base().ComponentID = c.componentID
	return attachable
}

// Type returns the reflect.Type of the component data stored in this column.
func (c *Column[T, PT]) Type() reflect.Type {
	return c.typeOfT
}

// Len returns the number of components currently stored in this column.
func (c *Column[T, PT]) Len() int {
	return c.Length
}

// Cap returns the total capacity of this column, which is the number of slots available for components.
func (c *Column[T, PT]) Cap() int {
	return len(c.data) * chunkSize
}

// ID returns the component ID associated with this column.
func (c *Column[T, PT]) ID() ComponentID {
	return c.componentID
}

// String returns a string representation of the component type stored in this column.
func (c *Column[T, PT]) String() string {
	return c.typeOfT.String()
}
