// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ecs

/*
Closed hash table, indexed directly by entities

	This is designed to be extremely fast due to the direct hashing, small
	number of elements, etc...
*/
type EntityTable []Entity

const EntityTableGrowthRate = 8

func (table *EntityTable) Set(entity Entity) {
	size := uint32(len(*table))
	if size == 0 {
		*table = make(EntityTable, 1)
		size = 1
	}
	// Fast path, common. Avoid doing modulus
	if size == 1 && (*table)[0] == 0 {
		(*table)[0] = entity
		return
	}
	i := uint32(entity) % size
	for range size {
		if (*table)[i] == 0 || (*table)[i] == entity {
			(*table)[i] = entity
			return
		}
		i = (i + 1) % size
	}

	// If we're here, that means we didn't find an empty slot.
	// We need to expand the table and rehash.
	newTable := make(EntityTable, size+EntityTableGrowthRate)
	for _, e := range *table {
		if e == 0 {
			continue
		}
		newTable.Set(e)
	}
	newTable.Set(entity)
	*table = newTable
}

// This method should be as efficient as possible, it gets called a lot!
func (table EntityTable) Contains(entity Entity) bool {
	size := uint32(len(table))
	if size == 0 {
		return false
	}
	if size == 1 && table[0] == entity {
		return true
	}
	i := uint32(entity) % size
	for range size {
		if table[i] == 0 {
			break
		} else if table[i] == entity {
			return true
		}
		i = (i + 1) % size
	}
	return false
}

func (table *EntityTable) Delete(entity Entity) bool {
	size := uint32(len(*table))
	if size == 0 {
		return false
	}
	// First, find our slot by linear probing
	i := uint32(entity) % size
	found := false
	for range size {
		if (*table)[i] == 0 {
			// Already nil, nothing to do
			return false
		}
		if (*table)[i] == entity {
			found = true
			break
		}
		i = (i + 1) % size
	}
	if !found {
		return false
	}

	// Erase current slot
	(*table)[i] = 0
	// Compact/rehash by moving non-nil elements in slots that don't match their
	// hash value.
	prev := i
	i = (i + 1) % size
	for range size {
		if (*table)[i] == 0 {
			return true
		}
		e := (*table)[i]
		hash := uint32(e) % size
		if hash != i {
			(*table)[prev], (*table)[i] = (*table)[i], 0
		} else {
			return true
		}
		prev = i
		i = (i + 1) % size
	}
	return true
}

func (table EntityTable) Serialize() []string {
	result := make([]string, 0)
	for _, e := range table {
		if e == 0 {
			continue
		}
		result = append(result, e.String())
	}
	return result
}

func (table EntityTable) String() string {
	result := ""
	for _, e := range table {
		if e == 0 {
			continue
		}
		if len(result) > 0 {
			result += ","
		}
		result += e.String()
	}
	return result
}

func (table EntityTable) First() Entity {
	for _, e := range table {
		if e != 0 {
			return e
		}
	}
	return 0
}

func (table EntityTable) Len() int {
	size := 0
	for _, e := range table {
		if e != 0 {
			size++
		}
	}
	return size
}
