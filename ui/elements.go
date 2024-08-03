// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ui

import "tlyakhov/gofoom/concepts"

type textElement struct {
	Rune    rune
	Color   *concepts.Vector4
	BGColor *concepts.Vector4
	Shadow  bool
}

type IWidget interface {
	GetWidget() *Widget
}

type Widget struct {
	Label     string
	highlight concepts.DynamicValue[concepts.Vector4]
}

type Page struct {
	Widgets      []IWidget
	SelectedItem int
}

type Button struct {
	Widget

	Clicked func(b *Button)
}

type Checkbox struct {
	Widget

	IsChecked bool

	Checked func(cb *Checkbox)
}

type Slider struct {
	Widget

	Value, Min, Max int
	Step            int

	Moved func(s *Slider)
}

func (e *Widget) GetWidget() *Widget {
	return e
}
