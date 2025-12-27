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

func (lc *LinkedController) ComponentID() ComponentID {
	return LinkedCID
}

func (lc *LinkedController) Methods() ControllerMethod {
	return ControllerPrecompute
}

func (lc *LinkedController) Target(target Component, e Entity) bool {
	lc.Entity = e
	lc.Linked = target.(*Linked)
	return lc.Linked.IsActive()
}

func (lc *LinkedController) Precompute() {
	// Remove this entity from any linked copies
	for _, c := range lc.SourceComponents {
		if c != nil {
			detach(c.ComponentID(), lc.Entity, false)
		}
	}
	lc.SourceComponents = make(ComponentTable, 0)
	for _, sourceEntity := range lc.Sources {
		for _, c := range AllComponents(sourceEntity) {
			if c == nil || !c.Shareable() {
				continue
			}
			existing := GetComponent(lc.Entity, c.ComponentID())
			if lc.AlwaysReplace && existing != nil {
				detach(c.ComponentID(), lc.Entity, false)
			} else if existing != nil {
				continue
			}
			lc.SourceComponents.Set(c)
			attach(lc.Entity, &c, c.ComponentID())
		}
	}
}
