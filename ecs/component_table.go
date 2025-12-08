// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ecs

import "sync/atomic"

/*
ComponentTable is a closed hash table that stores components, indexed by their
component IDs.
It is designed for high performance due to direct hashing and a small number of elements.
*/
type ComponentTable []Component

// ComponentTableGrowthRate is the rate at which the component table grows when
// it needs to be resized.
const ComponentTableGrowthRate = 8

// ComponentTableHit and ComponentTableMiss are atomic counters used for
// performance analysis of the component table.
var ComponentTableHit, ComponentTableMiss atomic.Uint64

// Set adds or updates a component in the table.
func (table *ComponentTable) Set(a Component) {
	cid := a.ComponentID()
	size := ComponentID(len(*table))
	if size == 0 {
		// If the table is empty, initialize it with the growth rate.
		*table = make(ComponentTable, ComponentTableGrowthRate)
		size = ComponentTableGrowthRate
	}
	// Calculate the initial index using the component ID modulo the table size.
	i := cid % size
	// Use linear probing to find an empty slot or a slot with the same component ID.
	for range size {
		if (*table)[i] == nil || (*table)[i].ComponentID() == cid {
			// Place the component in the found slot.
			(*table)[i] = a
			return
		}
		// Move to the next slot.
		i = (i + 1) % size
	}

	// If we're here, that means we didn't find an empty slot within the current
	// table size. We need to expand the table and rehash.
	newTable := make(ComponentTable, size+ComponentTableGrowthRate)
	// Copy existing components to the new table.
	for _, c := range *table {
		if c == nil {
			continue
		}
		newTable.Set(c)
	}
	// Add the new component to the expanded table.
	newTable.Set(a)
	// Replace the old table with the new one.
	*table = newTable
}

// Get retrieves a component from the table by its component ID.
// This method is performance-critical and should be as efficient as possible.
func (table ComponentTable) Get(cid ComponentID) Component {
	size := ComponentID(len(table))
	if size == 0 {
		return nil
	}
	// Calculate the initial index using the component ID modulo the table size.
	i := cid % size
	// Use linear probing to find the component.
	for range size {
		if table[i] == nil {
			// If an empty slot is encountered, the component is not in the table.
			break
		} else if table[i].ComponentID() == cid {
			// If a component with the matching ID is found, return it.
			//ComponentTableHit.Add(1)
			return table[i]
		}
		//ComponentTableMiss.Add(1)
		// Move to the next slot.
		i = (i + 1) % size
	}
	// If the component is not found after probing all slots, return nil.
	return nil
}

// Delete removes a component from the table by its component ID.
func (table *ComponentTable) Delete(cid ComponentID) {
	size := ComponentID(len(*table))
	if size == 0 {
		return
	}
	// First, find our slot by linear probing
	i := cid % size
	found := false
	for range size {
		if (*table)[i] == nil {
			// Already nil, nothing to do
			return
		}
		if (*table)[i].ComponentID() == cid {
			found = true
			break
		}
		i = (i + 1) % size
	}
	if !found {
		return
	}

	// Erase current slot
	(*table)[i] = nil

	// See https://en.wikipedia.org/wiki/Open_addressing
	// Compact/rehash by moving non-nil elements that are displaced into the
	// newly created hole.
	j := i
	for range size {
		j = (j + 1) % size
		if (*table)[j] == nil {
			return
		}
		cid := (*table)[j].ComponentID()
		hash := cid % size

		// Check if the element at j is displaced and should fill the hole at i.
		// The element at j should move to i if i is "between" hash and j cyclically.
		// That is, if i falls in the range [hash, j).
		if i <= j {
			// Normal case: hash <= j. Range is [hash, j).
			if i < hash && hash <= j {
				continue
			}
		} else {
			// Wrap-around case: hash > j. Range is [hash, size) U [0, j).
			if hash <= j || i < hash {
				continue
			}
		}

		(*table)[i] = (*table)[j]
		(*table)[j] = nil
		i = j
	}
}
