// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ecs

import "log"

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
	// Populate this table with the actual components from our source entities.
	lc.SourceComponents = make(ComponentTable, 0)
	for _, sourceEntity := range lc.Sources {
		for _, c := range AllComponents(sourceEntity) {
			if c == nil {
				continue
			}
			if c.ComponentID() == LinkedCID || c.ComponentID() == NamedCID {
				continue
			}
			// If this components isn't shareable and Linked.CopyUnshareable isn't
			// set, skip it.
			if !lc.CopyUnshareable && !c.Shareable() {
				continue
			}
			// Is there one already attached?
			existing := GetComponent(lc.Entity, c.ComponentID())
			// We have an existing component with the same ID as the one form
			// the source entity.
			if existing != nil {
				// Is it the same one?
				if existing == c || existing.Base().Entities.Contains(sourceEntity) {
					lc.SourceComponents.Set(existing)
					continue
				} else if existing.Base().Entities.Contains(sourceEntity) {
					log.Printf("ecs.LinkedController: warning - found existing component %v on entity %v that is shared with source entity %v but is NOT the same pointer.", existing, lc.Entity, sourceEntity)
					lc.SourceComponents.Set(existing)
					continue
				}
				if lc.AlwaysReplace {
					// Replace whatever this entity used to have.
					detach(c.ComponentID(), lc.Entity, false)
				} else {
					// Skip this
					continue
				}
			}
			if c.Shareable() {
				// Only set this for linked (not copied) components.
				lc.SourceComponents.Set(c)
				attach(lc.Entity, &c, c.ComponentID())
			} else {
				cloned := LoadComponentWithoutAttaching(c.ComponentID(), c.Serialize())
				ModifyComponentRelationEntities(cloned, func(r *Relation, ref Entity) Entity {
					// Update self-references to the new entity
					if ref == sourceEntity {
						return lc.Entity
					}
					return ref
				})
				attach(lc.Entity, &cloned, cloned.ComponentID())
			}
		}
	}
}
