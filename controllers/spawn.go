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

func CloneEntity(e ecs.Entity, onCloneComponent func(e ecs.Entity, cid ecs.ComponentID, copiedComponent ecs.Component, pastedComponent ecs.Component) bool) ecs.Entity {
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
		ecs.ActAllControllersOneEntity(pastedEntity, ecs.ControllerRecalculate)
		ecs.ActAllControllersOneEntity(pastedEntity, ecs.ControllerFrame)
		return pastedEntity
	}

	// We didn't attach anything
	return 0
}

func RespawnInventory(c *inventory.Carrier) {
	// We have a copy of the spawn entity's carrier. Let's copy the slots over as
	// well.
	cloned := ecs.EntityTable{}
	mapped := map[ecs.Entity]ecs.Entity{}
	for _, copied := range c.Slots {
		if copied == 0 {
			continue
		}
		pasted := CloneEntity(copied, func(e ecs.Entity, cid ecs.ComponentID, original ecs.Component, pasted ecs.Component) bool {
			if cid == inventory.SlotCID {
				pasted.Base().Flags |= ecs.ComponentHideEntityInEditor
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

func Respawn(s *behaviors.Spawner) {
	// First, delete previously spawned entities
	for e := range s.Spawned {
		// Need to delete old inventory slots
		if carrier := inventory.GetCarrier(e); carrier != nil {
			for _, slot := range carrier.Slots {
				// Don't need to check for zero, Delete will ignore it.
				ecs.Delete(slot)
			}
			carrier.Slots = ecs.EntityTable{}
		}
		ecs.Delete(e)
	}
	s.Spawned = make(map[ecs.Entity]int64)

	// Pick a random spawner
	randomSpawner := s.Entity
	if s.Entities.Len() > 0 {
		picked := rand.Int() % s.Entities.Len()
		i := 0
		for _, e := range s.Entities {
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
		return
	}

	pastedEntity := CloneEntity(randomSpawner, func(e ecs.Entity, cid ecs.ComponentID, _ ecs.Component, pasted ecs.Component) bool {
		pasted.Base().Flags |= ecs.ComponentHideEntityInEditor
		switch cid {
		case behaviors.SpawnerCID:
			return false
		case ecs.LinkedCID:
			// Don't copy over linked components, we shouldn't tie anything
			// to the newly spawned entity.
			return false
		case ecs.NamedCID:
			named := pasted.(*ecs.Named)
			named.Name = fmt.Sprintf("%v (spawned from %v)", named.Name, randomSpawner.ShortString())
		case inventory.CarrierCID:
			carrier := pasted.(*inventory.Carrier)
			RespawnInventory(carrier)
		}
		return true
	})

	if pastedEntity != 0 {
		s.Spawned[pastedEntity] = ecs.Simulation.SimTimestamp
		spawnee := ecs.NewAttachedComponent(pastedEntity, behaviors.SpawneeCID).(*behaviors.Spawnee)
		spawnee.Flags = ecs.ComponentActive | ecs.ComponentLockedInEditor | ecs.ComponentHideEntityInEditor
	}
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
		Respawn(s)
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
