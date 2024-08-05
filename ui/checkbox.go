// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ui

type Checkbox struct {
	Widget

	Value bool

	Checked func(cb *Checkbox)
}

func (cb *Checkbox) Serialize() map[string]any {
	result := cb.Widget.Serialize()
	result["Value"] = cb.Value
	return result
}

func (cb *Checkbox) Construct(data map[string]any) {
	if v, ok := data["Value"]; ok {
		cb.Value = v.(bool)
	}
}
