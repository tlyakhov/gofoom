// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ecs

// BaseController is a base implementation for controllers. It provides default
// implementations for controller methods. Controllers should embed this struct
// and override the methods they need to implement.
type BaseController struct {
	// Entity is the entity ID associated with this controller.
	Entity
}

// ComponentID returns the component ID that this controller operates on. The
// base implementation returns 0, which should be overridden by concrete controllers.
func (c *BaseController) ComponentID() ComponentID {
	return 0
}

// Methods returns the controller methods that this controller implements. The
// base implementation returns 0, which should be overridden.
func (c *BaseController) Methods() ControllerMethod {
	return 0
}

// EditorPausedMethods returns the controller methods that this controller
// implements when the editor is paused. The base implementation returns 0.
func (c *BaseController) EditorPausedMethods() ControllerMethod {
	return 0
}

// Target determines whether the controller should act on a specific entity and
// component. The base implementation always returns false.
func (c *BaseController) Target(a Component, entity Entity) bool {
	return false
}

// Frame is a controller method that is called every tick. The base
// implementation does nothing.
func (c *BaseController) Frame() {}

// Recalculate is a controller method that is called when a component is
// attached or detached, or when a linked component changes. The base implementation does nothing.
func (c *BaseController) Recalculate() {}
