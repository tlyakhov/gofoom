// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ecs

// Closed hash table, indexed directly by component indices
// This is designed to be extremely fast due to the direct hashing, small
// number of elements, etc...
type ComponentTable []Attachable

const ComponentTableGrowthRate = 8

func (table *ComponentTable) Set(a Attachable) {
	cid := a.GetComponentID()
	size := len(*table)
	if size == 0 {
		*table = make(ComponentTable, ComponentTableGrowthRate)
		size = ComponentTableGrowthRate
	}
	i := int(cid>>16) % size
	for range size {
		if (*table)[i] == nil || (*table)[i].GetComponentID() == cid {
			(*table)[i] = a
			return
		}
		i = (i + 1) % size
	}

	// Rehash every few attached components
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
	size := len(table)
	if size == 0 {
		return nil
	}
	i := int(cid>>16) % size
	for range size {
		if table[i] == nil || table[i].GetComponentID() == cid {
			return table[i]
		}
		i = (i + 1) % size
	}
	return nil
}

func (table *ComponentTable) Delete(cid ComponentID) {
	size := len(*table)
	if size == 0 {
		return
	}
	// First, find our slot by linear probing
	i := (int(cid >> 16)) % size
	found := false
	for range size {
		if (*table)[i] == nil {
			// Already nil, nothing to do
			return
		}
		if (*table)[i].GetComponentID() == cid {
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
	for range size {
		if (*table)[i] == nil {
			return
		}
		cid := (*table)[i].GetComponentID()
		hash := int(cid>>16) % size
		if hash != i {
			(*table)[prev], (*table)[i] = (*table)[i], nil
		} else {
			return
		}
		prev = i
		i = (i + 1) % size
	}
}
