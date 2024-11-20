// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ecs

type LinkedController struct {
	BaseController
	*Linked
}

func init() {
	// Should run before everything
	Types().RegisterController(func() Controller { return &LinkedController{} }, 0)
}

func (ic *LinkedController) ComponentID() ComponentID {
	return LinkedCID
}

func (ic *LinkedController) Methods() ControllerMethod {
	return ControllerRecalculate
}

func (ic *LinkedController) Target(target Attachable, e Entity) bool {
	ic.Entity = e
	ic.Linked = target.(*Linked)
	return ic.Linked.IsActive()
}

func (ic *LinkedController) Recalculate() {
	ic.Linked.Recalculate()
}
