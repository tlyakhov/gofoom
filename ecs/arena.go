// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"log"
	"reflect"

	"github.com/kelindar/bitmap"
)

// chunkSize is the size of each chunk in the arena. Arenas are split into
// chunks to improve memory management and reduce the cost of resizing.
const chunkSize = 64

// componentChunk represents a fixed-size array of components.
// Note: we can't use a slice for the data, because we use long-lived pointers to
// components, and slices can reallocate the backing array when resized,
// invalidating existing pointers.
// See https://utcc.utoronto.ca/~cks/space/blog/programming/GoSlicesVsPointers
type componentChunk[T any, PT GenericAttachable[T]] [chunkSize]T

// Arena is a sparse storage for components of a specific type.
// It stores components in chunks of fixed size, allowing for efficient access
// and iteration.
// The type parameters are:
//   - T: The type of the component struct.
//   - PT: A pointer to the component type, which must implement the
//     GenericAttachable interface.
type Arena[T any, PT GenericAttachable[T]] struct {
	// Length is the number of components currently stored in the arena.
	Length int
	// Getter is a function that retrieves a component of this type for a given entity.
	Getter func(e Entity) PT

	data []*componentChunk[T, PT]
	// fill is a bitmap that tracks which slots in the arena are occupied by components.
	fill bitmap.Bitmap
	// typeOfT is the reflect.Type of the component data.
	typeOfT reflect.Type
	// componentID is the unique identifier for the component type this arena stores.
	componentID ComponentID
}

// From initializes a arena from another arena of the same type, copying
// metadata and setting the Universe instance.
func (col *Arena[T, PT]) From(source AttachableArena) {
	placeholder := source.(*Arena[T, PT])
	col.typeOfT = placeholder.typeOfT
	col.componentID = placeholder.componentID
	col.Getter = placeholder.Getter
}

// Value retrieves the component at the given index in the arena.
// It performs no bounds checking for performance reasons.
func (col *Arena[T, PT]) Value(index int) PT {
	// No bounds checking for performance. This should always be inlined
	ptr := PT(&(col.data[index/chunkSize][index%chunkSize]))
	if !ptr.IsAttached() {
		return nil
	}
	return ptr
}

// Attachable retrieves the component at the given index as an Attachable interface.
// It performs no bounds checking for performance reasons.
func (col *Arena[T, PT]) Attachable(index int) Attachable {
	// No bounds checking for performance. This should always be inlined
	// Duplicates code in .Value() because the return type is different here and nil
	// in golang behaves idiosyncratically
	ptr := PT(&(col.data[index/chunkSize][index%chunkSize]))
	if !ptr.IsAttached() {
		return nil
	}
	return ptr
}

// Detach removes the component at the given index from the arena.
// It uses a fill bitmap to mark the slot as empty instead of shifting elements.
func (col *Arena[T, PT]) Detach(index int) {
	if index >= col.Cap() {
		log.Printf("ecs.Arena.Detach: found component index %v, but component list is too short.", index)
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

// AddTyped adds a component to the arena, automatically handling the Attachable
// interface conversion.
func (col *Arena[T, PT]) AddTyped(component *PT) {
	attachable := Attachable(*component)
	col.Add(&attachable)
	*component = attachable.(PT)
}

// Add adds a component to the arena at the next available slot.
// It uses a fill bitmap to find the next free index.
func (col *Arena[T, PT]) Add(component *Attachable) {
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
	(*component).Base().indexInArena = (int(nextFree))
	col.Length++
}

// Replace replaces the component at the given index with the provided
// component. Importantly, the argument is a pointer - it will be replaced with
// the resulting memory location.
func (col *Arena[T, PT]) Replace(component *Attachable, index int) {
	if *component == nil {
		*component = col.Value(index)
		return
	}
	ptr := col.Value(index)
	*ptr = *((*component).(PT))
	*component = ptr
	ptr.Base().indexInArena = index
}

// New creates a new Attachable component of the type stored in this arena.
func (c *Arena[T, PT]) New() Attachable {
	var component T
	attachable := PT(&component)
	return attachable
}

// Type returns the reflect.Type of the component data stored in this arena.
func (c *Arena[T, PT]) Type() reflect.Type {
	return c.typeOfT
}

// Len returns the number of components currently stored in this arena.
func (c *Arena[T, PT]) Len() int {
	return c.Length
}

// Cap returns the total capacity of this arena, which is the number of slots available for components.
func (c *Arena[T, PT]) Cap() int {
	return len(c.data) * chunkSize
}

// ID returns the component ID associated with this arena.
func (c *Arena[T, PT]) ID() ComponentID {
	return c.componentID
}

// String returns the component type stored in this arena.
func (c *Arena[T, PT]) String() string {
	return c.typeOfT.String()
}
