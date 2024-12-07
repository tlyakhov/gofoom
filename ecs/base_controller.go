// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ecs

type BaseController struct {
	Entity
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

func (c *BaseController) Target(a Attachable, entity Entity) bool {
	return false
}

func (c *BaseController) Always()      {}
func (c *BaseController) Recalculate() {}
