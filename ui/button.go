// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ui

type Button struct {
	Widget

	Clicked func(b *Button)
}
