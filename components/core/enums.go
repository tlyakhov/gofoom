// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package core

type CollisionResponse int

//go:generate go run github.com/dmarkham/enumer -type=CollisionResponse -json
const (
	CollideNone CollisionResponse = iota
	CollideDeactivate
	CollideSeparate
	CollideBounce
	CollideStop
	CollideRemove
)

//go:generate go run github.com/dmarkham/enumer -type=BodyShadow -json
type BodyShadow int

const (
	BodyShadowNone BodyShadow = iota
	BodyShadowImage
	BodyShadowSphere
	BodyShadowAABB
)

//go:generate go run github.com/dmarkham/enumer -type=ScriptStyle -json
type ScriptStyle int

const (
	ScriptStyleRaw ScriptStyle = iota
	ScriptStyleBoolExpr
	ScriptStyleStatement
)
