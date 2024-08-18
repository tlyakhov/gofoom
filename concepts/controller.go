// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package concepts

type ControllerMethod uint32

const (
	ControllerAlways ControllerMethod = 1 << iota
	ControllerLoaded
	ControllerRecalculate
)

type Controller interface {
	ComponentIndex() int
	Methods() ControllerMethod
	// Return false if controller shouldn't run for this entity
	Target(Attachable) bool
	Always()
	Loaded()
	Recalculate()
}

type BaseController struct {
}

func (c *BaseController) ComponentIndex() int {
	return AttachedComponentIndex
}

func (c *BaseController) Methods() ControllerMethod {
	return 0
}

func (c *BaseController) Target(a Attachable) bool {
	return false
}

func (c *BaseController) Always()      {}
func (c *BaseController) Loaded()      {}
func (c *BaseController) Recalculate() {}
