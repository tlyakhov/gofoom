// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ui

type Slider struct {
	Widget

	Value, Min, Max int
	Step            int

	Moved func(s *Slider)
}

func (s *Slider) Serialize() map[string]any {
	result := s.Widget.Serialize()
	result["Value"] = s.Value
	return result
}
