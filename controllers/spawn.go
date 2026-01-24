// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"fmt"
	"log"
	"math/rand"
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/inventory"
	"tlyakhov/gofoom/ecs"
)

type cloneComponentFunc func(e ecs.Entity, cid ecs.ComponentID, copiedComponent ecs.Component, pastedComponent ecs.Component) bool

func CloneEntity(e ecs.Entity, preserveLinks bool, onCloneComponent cloneComponentFunc) ecs.Entity {
	// TODO: This kind of cloning operation is used in other places (e.g. the
	// editor). Should this be pulled into ecs? Will need to figure out how to
	// address deep vs. shallow cloning and wiring up any relationships.
	copiedEntityData := ecs.SerializeEntity(e, false)
	pastedEntity := ecs.NewEntity()
	var originalComponent ecs.Component
	var pastedComponentData map[string]any
	for name, cid := range ecs.Types().IDs {
		copiedData := copiedEntityData[name]
		if copiedData == nil {
			continue
		}
		// This component is a reference to another entity.
		if ref, ok := copiedData.(string); ok {
			refEntity, err := ecs.ParseEntity(ref)
			if err != nil {
				log.Printf("controllers.CloneEntity: error parsing referenced entity %v", ref)
				continue
			}
			originalComponent = ecs.GetComponent(refEntity, cid)
			if preserveLinks {
				ecs.Attach(cid, pastedEntity, &originalComponent)
				continue
			}
			pastedComponentData = originalComponent.Serialize()
		} else {
			originalComponent = ecs.GetComponent(e, cid)
			pastedComponentData = copiedData.(map[string]any)
		}
		pastedComponent := ecs.LoadComponentWithoutAttaching(cid, pastedComponentData)
		if onCloneComponent == nil || onCloneComponent(pastedEntity, cid, originalComponent, pastedComponent) {
			ecs.ModifyComponentRelationEntities(pastedComponent, func(r *ecs.Relation, ref ecs.Entity) ecs.Entity {
				// Update self-references to the new entity
				if ref == e {
					return pastedEntity
				}
				return ref
			})
			ecs.Attach(cid, pastedEntity, &pastedComponent)
		}
	}
	for _, c := range ecs.AllComponents(pastedEntity) {
		if c == nil {
			continue
		}
		// We attached at least one component to the new entity.
		ecs.ActAllControllersOneEntity(pastedEntity, ecs.ControllerPrecompute)
		ecs.ActAllControllersOneEntity(pastedEntity, ecs.ControllerFrame)
		return pastedEntity
	}

	// We didn't attach anything
	return 0
}

func spawnInventory(c *inventory.Carrier) {
	// We have a copy of the spawn entity's carrier. Let's copy the slots over as
	// well.
	cloned := ecs.EntityTable{}
	mapped := map[ecs.Entity]ecs.Entity{}
	for _, copied := range c.Slots {
		if copied == 0 {
			continue
		}
		pasted := CloneEntity(copied, false, func(e ecs.Entity, cid ecs.ComponentID, original ecs.Component, pasted ecs.Component) bool {
			if cid == inventory.SlotCID {
				pasted.Base().Flags |= ecs.ComponentHideEntityInEditor
				pasted.(*inventory.Slot).Count.ResetToSpawn()
				return true
			}
			// Otherwise, just attach the original.
			ecs.Attach(cid, e, &original)
			return false
		})
		if pasted != 0 {
			cloned.Set(pasted)
			mapped[copied] = pasted
		}
	}
	c.Slots = cloned
	if c.SelectedWeapon != 0 && mapped[c.SelectedWeapon] != 0 {
		c.SelectedWeapon = mapped[c.SelectedWeapon]
	} else {
		c.SelectedWeapon = 0
	}
}

func DeleteSpawned(s *behaviors.Spawner) {
	for e := range s.Spawned {
		// Validate that we have a link between Spawner and Spawnee.
		if spawnee := behaviors.GetSpawnee(e); spawnee != nil && spawnee.Spawner == s.Entity {
			ecs.Delete(e)
		}
	}
	s.Spawned = make(map[ecs.Entity]int64)
}

func Spawn(s *behaviors.Spawner) ecs.Entity {
	// Pick a random spawner
	randomSpawner := s.Entity
	if s.Targets.Len() > 0 {
		picked := rand.Int() % s.Targets.Len()
		i := 0
		for _, e := range s.Targets {
			if e == 0 {
				continue
			}
			if i == picked {
				randomSpawner = e
				break
			}
			i++
		}
	}
	if randomSpawner == 0 {
		log.Printf("Error, no spawners found for %v", s.Entity)
		return 0
	}

	pastedEntity := CloneEntity(randomSpawner, s.PreserveLinks,
		func(e ecs.Entity, cid ecs.ComponentID, _ ecs.Component, pasted ecs.Component) bool {
			pasted.Base().Flags |= ecs.ComponentHideEntityInEditor | ecs.ComponentActive
			switch cid {
			case behaviors.SpawnerCID:
				// Don't copy over spawners
				return false
			case ecs.LinkedCID:
				// By default, don't copy over linked components, we want spawned
				// entities to have their own lifecycle without modifying the spawner.
				return s.PreserveLinks
			case ecs.NamedCID:
				named := pasted.(*ecs.Named)
				named.Name = fmt.Sprintf("%v (spawned from %v)", named.Name, randomSpawner.ShortString())
			case inventory.CarrierCID:
				carrier := pasted.(*inventory.Carrier)
				spawnInventory(carrier)
			}
			return true
		})

	if pastedEntity != 0 {
		s.Spawned[pastedEntity] = ecs.Simulation.SimTimestamp
		spawnee := ecs.NewAttachedComponent(pastedEntity, behaviors.SpawneeCID).(*behaviors.Spawnee)
		spawnee.Flags = ecs.ComponentActive | ecs.ComponentLockedInEditor | ecs.ComponentHideEntityInEditor
		spawnee.Spawner = s.Entity
	}
	return pastedEntity
}

func RespawnAll() {
	// TODO: move this out of here
	core.QuadTree.Reset()
	arena := ecs.ArenaFor[behaviors.Spawner](behaviors.SpawnerCID)
	for i := range arena.Cap() {
		s := arena.Value(i)
		if s == nil {
			continue
		}
		if s.Auto != behaviors.AutoSpawnNone {
			DeleteSpawned(s)
		}
		if s.Auto == behaviors.AutoSpawnOnLoad {
			Spawn(s)
		}
	}
}

func ResetAllSpawnables() {
	for d := range ecs.Simulation.Spawnables {
		d.ResetToSpawn()
	}

	for d := range ecs.Simulation.Dynamics {
		if a := d.GetAnimation(); a != nil {
			a.Reset()
		}
	}
}
