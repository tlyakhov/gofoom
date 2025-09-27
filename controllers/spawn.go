// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"log"
	"math/rand"
	"tlyakhov/gofoom/components/character"
	"tlyakhov/gofoom/components/inventory"
	"tlyakhov/gofoom/ecs"
)

func Respawn(force bool) {
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
		ecs.Delete(players[len(players)-1].Entity)
		players = players[:len(players)-1]
	}

	if len(players) > 0 || len(spawns) == 0 {
		return
	}

	spawn := spawns[rand.Int()%len(spawns)]
	// TODO: This kind of cloning operation is used in other places (e.g. the
	// editor). Should this be pulled into ecs? Will need to figure out how to
	// address deep vs. shallow cloning and wiring up any relationships.
	copiedSpawn := ecs.SerializeEntity(spawn.Entity)
	var pastedEntity ecs.Entity
	var mappedComponent map[string]any
	for name, cid := range ecs.Types().IDs {
		mappedData := copiedSpawn[name]
		if mappedData == nil {
			continue
		}
		if pastedEntity == 0 {
			pastedEntity = ecs.NewEntity()
		}
		// This component is a reference to another entity.
		if refData, ok := mappedData.(string); ok {
			eRef, err := ecs.ParseEntity(refData)
			if err != nil {
				log.Printf("controllers.Respawn: error parsing referenced entity %v", refData)
				continue
			}
			mappedComponent = ecs.GetComponent(eRef, cid).Serialize()
		} else {
			mappedComponent = mappedData.(map[string]any)
		}
		c := ecs.LoadComponentWithoutAttaching(cid, mappedComponent)
		ecs.Attach(cid, pastedEntity, &c)
		switch cid {
		case character.PlayerCID:
			player := c.(*character.Player)
			player.Spawn = false
		case ecs.NamedCID:
			named := c.(*ecs.Named)
			named.Name = "Player"
		case inventory.CarrierCID:
			// carrier := c.(*inventory.Carrier)
			// TODO: Clone/respawn inventory
		}
	}
	ecs.ActAllControllersOneEntity(pastedEntity, ecs.ControllerRecalculate)
	ecs.ActAllControllersOneEntity(pastedEntity, ecs.ControllerAlways)
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
