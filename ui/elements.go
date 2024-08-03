// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ui

import "tlyakhov/gofoom/concepts"

type IWidget interface {
	GetWidget() *Widget
}

type Widget struct {
	Label  string
	Action func()
}

type Page struct {
	Items        []IWidget
	SelectedItem int
}

type Button struct {
	Widget

	bgColor concepts.DynamicValue[concepts.Vector4]
}

func (e *Widget) GetWidget() *Widget {
	return e
}
