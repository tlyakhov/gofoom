// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ecs

type ControllerMethod uint32

const (
	ControllerAlways ControllerMethod = 1 << iota
	ControllerLoaded
	ControllerRecalculate
)

type Controller interface {
	ComponentID() ComponentID
	Methods() ControllerMethod
	EditorPausedMethods() ControllerMethod
	// Return false if controller shouldn't run for this entity
	Target(Attachable) bool
	Always()
	Loaded()
	Recalculate()
}

type BaseController struct {
}

func (c *BaseController) ComponentID() ComponentID {
	return 0
}

func (c *BaseController) Methods() ControllerMethod {
	return 0
}

func (c *BaseController) EditorPausedMethods() ControllerMethod {
	return 0
}

func (c *BaseController) Target(a Attachable) bool {
	return false
}

func (c *BaseController) Always()      {}
func (c *BaseController) Loaded()      {}
func (c *BaseController) Recalculate() {}
