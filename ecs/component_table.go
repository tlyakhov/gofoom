// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ecs

/*
Closed hash table, indexed directly by component indices

	This is designed to be extremely fast due to the direct hashing, small
	number of elements, etc...

	It handles "instances" as a special case, recursively querying parent tables
	if the entity in question has an Linked component attached.
*/
type ComponentTable []Attachable

const ComponentTableGrowthRate = 8

func (table *ComponentTable) Set(a Attachable) {
	cid := a.Base().ComponentID
	size := len(*table)
	if size == 0 {
		*table = make(ComponentTable, ComponentTableGrowthRate)
		size = ComponentTableGrowthRate
	}
	// 0 is reserved for Linked components
	if cid == LinkedCID {
		(*table)[0] = a
		return
	}
	i := 1 + int(cid>>16)%(size-1)
	for range size - 1 {
		if i == 0 {
			i = 1
		}
		if (*table)[i] == nil || (*table)[i].Base().ComponentID == cid {
			(*table)[i] = a
			return
		}
		i = (i + 1) % size
	}

	// If we're here, that means we didn't find an empty slot.
	// We need to expand the table and rehash.
	newTable := make(ComponentTable, size+ComponentTableGrowthRate)
	for _, c := range *table {
		if c == nil {
			continue
		}
		newTable.Set(c)
	}
	newTable.Set(a)
	*table = newTable
}

// This method should be as efficient as possible, it gets called a lot!
func (table ComponentTable) Get(cid ComponentID) Attachable {
	size := uint16(len(table))
	if size == 0 {
		return nil
	}
	// 0 is reserved for Linked components
	if cid == LinkedCID {
		return table[0]
	}
	i := 1 + uint16(cid>>16)%(size-1)
	for range size - 1 {
		if i == 0 {
			i = 1
		}
		if table[i] == nil {
			break
		} else if table[i].Base().ComponentID == cid {
			return table[i]
		}
		i = (i + 1) % size
	}
	// If this table isn't an "instance", just return nil, otherwise try the instance.
	if table[0] == nil {
		return nil
	}
	return table[0].(*Linked).SourceComponents.Get(cid)
}

func (table *ComponentTable) Delete(cid ComponentID) {
	size := uint16(len(*table))
	if size == 0 {
		return
	}
	// 0 is reserved for Linked components
	if cid == LinkedCID {
		(*table)[0] = nil
		return
	}
	// First, find our slot by linear probing
	i := 1 + (uint16(cid>>16))%(size-1)
	found := false
	for range size - 1 {
		if i == 0 {
			i = 1
		}
		if (*table)[i] == nil {
			// Already nil, nothing to do
			return
		}
		if (*table)[i].Base().ComponentID == cid {
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
	// Compact/rehash by moving non-nil elements in slots that don't match their
	// hash value.
	prev := i
	i = (i + 1) % size
	for range size - 1 {
		if i == 0 {
			i = 1
		}
		if (*table)[i] == nil {
			return
		}
		cid := (*table)[i].Base().ComponentID
		hash := 1 + uint16(cid>>16)%(size-1)
		if hash != i {
			(*table)[prev], (*table)[i] = (*table)[i], nil
		} else {
			return
		}
		prev = i
		i = (i + 1) % size
	}
}
