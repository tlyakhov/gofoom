// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package containers

import "fmt"

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

func (s Set[T]) First() T {
	for element := range s {
		return element
	}
	var empty T
	return empty
}

func (s Set[T]) Contains(element T) bool {
	_, ok := s[element]
	return ok
}

func (s Set[T]) String() string {
	r := ""
	for element := range s {
		if len(r) != 0 {
			r += ","
		}
		switch value := any(element).(type) {
		case string:
			r += value
		case fmt.Stringer:
			r += value.String()
		}
	}
	return r
}
