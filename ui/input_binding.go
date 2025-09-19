// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ui

import (
	"github.com/spf13/cast"
)

type InputBinding struct {
	Widget

	Input1   string
	Input2   string
	Selected int

	Changed func(s *InputBinding)
}

func (binding *InputBinding) Serialize() map[string]any {
	result := binding.Widget.Serialize()
	result["Input1"] = binding.Input1
	result["Input2"] = binding.Input2
	return result
}

func (binding *InputBinding) Construct(data map[string]any) {
	if v, ok := data["Input1"]; ok {
		binding.Input1 = cast.ToString(v)
	}
	if v, ok := data["Input2"]; ok {
		binding.Input2 = cast.ToString(v)
	}
}

func (ui *UI) inputBindingLabel(binding *InputBinding) string {
	label := binding.Label + " ["
	if binding.Input1 != "" {
		label += binding.Input1
	} else {
		label += " "
	}
	label += "] ["
	if binding.Input2 != "" {
		label += binding.Input2
	} else {
		label += " "
	}
	label += "]"
	return label
}

func (ui *UI) measureInputBinding(binding *InputBinding) (int, int) {
	return ui.measureBox(ui.inputBindingLabel(binding))
}

func (ui *UI) renderInputBinding(binding *InputBinding, x, y int) {

	hStart := ui.Padding + len(binding.Label) + 1
	hEnd := hStart + max(len(binding.Input1), 1) + 2
	if binding.Selected == 1 {
		hStart = hEnd + 1
		hEnd = hStart + max(len(binding.Input2), 1) + 2
	}
	ui.renderBox(&binding.Widget, ui.inputBindingLabel(binding), x, y, hStart, hEnd)
}
