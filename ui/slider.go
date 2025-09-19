// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ui

import (
	"strconv"
)

type Slider struct {
	Widget

	Value, Min, Max int
	Step            int

	Moved func(s *Slider)
}

func (s *Slider) Serialize() map[string]any {
	result := s.Widget.Serialize()
	result["Value"] = strconv.FormatInt(int64(s.Value), 10)
	return result
}

func (s *Slider) Construct(data map[string]any) {
	if v, ok := data["Value"]; ok {
		if v2, err := strconv.ParseInt(v.(string), 10, 64); err == nil {
			s.Value = max(min(int(v2), s.Max), s.Min)
		}
	}
}

func (ui *UI) measureSlider(s *Slider) (int, int) {
	label := s.Label + " [ " + strconv.Itoa(s.Value) + string(rune(29)) + "]"

	return ui.measureBox(label)
}

func (ui *UI) renderSlider(s *Slider, x, y int) {
	label := s.Label + " [ " + strconv.Itoa(s.Value) + string(rune(29)) + "]"

	ui.renderBox(&s.Widget, label, x, y, ui.Padding+len(s.Label)+1, ui.Padding+len(label))
}
