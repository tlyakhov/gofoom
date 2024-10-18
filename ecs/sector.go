// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ecs

type InstancedController struct {
	BaseController
	*Instanced
}

func init() {
	// Should run before everything
	Types().RegisterController(func() Controller { return &InstancedController{} }, 0)
}

func (ic *InstancedController) ComponentID() ComponentID {
	return InstancedCID
}

func (ic *InstancedController) Methods() ControllerMethod {
	return ControllerRecalculate
}

func (ic *InstancedController) Target(target Attachable) bool {
	ic.Instanced = target.(*Instanced)
	return ic.Instanced.IsActive()
}

func (ic *InstancedController) Recalculate() {
	ic.Instanced.Recalculate()
}
