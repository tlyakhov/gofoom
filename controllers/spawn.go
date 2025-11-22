// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"log"
	"math/rand"
	"tlyakhov/gofoom/components/character"
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

func Respawn(force bool) {
	core.QuadTree.Reset()

	spawns := make([]*character.Player, 0)
	players := make([]*character.Player, 0)
	arena := ecs.ArenaFor[character.Player](character.PlayerCID)
	for i := range arena.Cap() {
		p := arena.Value(i)
		if p == nil || !p.IsActive() {
			continue
		}
		if p.Spawn {
			spawns = append(spawns, p)
		} else {
			players = append(players, p)
		}
	}

	// Remove extra players
	// By default, avoid spawning a player if one exists
	maxPlayers := 1
	if force {
		maxPlayers = 0
	}

	for len(players) > maxPlayers {
		player := players[len(players)-1]
		// Need to delete old inventory slots
		if carrier := inventory.GetCarrier(player.Entity); carrier != nil {
			for _, slot := range carrier.Slots {
				// Don't need to check for zero, Delete will ignore it.
				ecs.Delete(slot)
			}
			carrier.Slots = ecs.EntityTable{}
		}
		ecs.Delete(player.Entity)
		players = players[:len(players)-1]
	}

	if len(players) > 0 || len(spawns) == 0 {
		return
	}

	spawn := spawns[rand.Int()%len(spawns)]

	CloneEntity(spawn.Entity, func(e ecs.Entity, cid ecs.ComponentID, _ ecs.Component, pasted ecs.Component) bool {
		switch cid {
		case ecs.LinkedCID:
			// Don't copy over linked components, we shouldn't tie anything
			// to the newly spawned player.
			return false
		case character.PlayerCID:
			player := pasted.(*character.Player)
			player.Spawn = false
			player.Flags |= ecs.ComponentHideEntityInEditor
		case ecs.NamedCID:
			named := pasted.(*ecs.Named)
			named.Name = "Player"
		case inventory.CarrierCID:
			carrier := pasted.(*inventory.Carrier)
			RespawnInventory(carrier)
		}
		return true
	})
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
