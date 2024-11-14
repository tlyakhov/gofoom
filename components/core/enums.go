// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package core

type CollisionResponse int

//go:generate go run github.com/dmarkham/enumer -type=CollisionResponse -json
const (
	CollideNone CollisionResponse = 1 << iota
	CollideDeactivate
	CollideSeparate
	CollideBounce
	CollideStop
	CollideRemove
)
