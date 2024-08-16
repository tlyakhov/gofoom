// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package concepts

type Set[T comparable] map[T]struct{}

func (s Set[T]) Add(element T) {
	s[element] = struct{}{}
}

func (s Set[T]) AddAll(elements ...T) {
	for _, element := range elements {
		s[element] = struct{}{}
	}
}

func (s Set[T]) Delete(element T) {
	delete(s, element)
}

func (s Set[T]) Contains(element T) bool {
	_, ok := s[element]
	return ok
}
