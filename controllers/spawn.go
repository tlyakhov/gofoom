// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"math/rand"
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/dynamic"
	"tlyakhov/gofoom/ecs"
)

func Respawn(u *ecs.Universe, force bool) {
	spawns := make([]*behaviors.Player, 0)
	players := make([]*behaviors.Player, 0)
	col := ecs.ColumnFor[behaviors.Player](u, behaviors.PlayerCID)
	for i := range col.Cap() {
		p := col.Value(i)
		if p == nil || !p.Active {
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
		u.Delete(players[len(players)-1].Entity)
		players = players[:len(players)-1]
	}

	if len(players) > 0 || len(spawns) == 0 {
		return
	}

	spawn := spawns[rand.Int()%len(spawns)]
	// TODO: This kind of cloning operation is used in other places (e.g. the
	// editor). Should this be pulled into Universe? Will need to figure out how to
	// address deep vs. shallow cloning and wiring up any relationships.
	copiedSpawn := u.SerializeEntity(spawn.Entity)
	var pastedEntity ecs.Entity
	for name, cid := range ecs.Types().IDs {
		mappedData := copiedSpawn[name]
		if mappedData == nil {
			continue
		}
		if pastedEntity == 0 {
			pastedEntity = u.NewEntity()
		}
		mappedComponent := mappedData.(map[string]any)
		c := u.LoadComponentWithoutAttaching(cid, mappedComponent)
		u.Attach(cid, pastedEntity, &c)
		if cid == behaviors.PlayerCID {
			player := c.(*behaviors.Player)
			player.Spawn = false
		} else if cid == ecs.NamedCID {
			named := c.(*ecs.Named)
			named.Name = "Player"
		} else if cid == behaviors.InventoryCarrierCID {
			// carrier := c.(*behaviors.InventoryCarrier)
			// TODO: Clone/respawn inventory
		}
	}
	u.ActAllControllersOneEntity(pastedEntity, ecs.ControllerRecalculate)
	u.ActAllControllersOneEntity(pastedEntity, ecs.ControllerAlways)
}

func ResetAllSpawnables(ecs *ecs.Universe) {
	ecs.Simulation.Spawnables.Range(func(d dynamic.Spawnable, _ struct{}) bool {
		d.ResetToSpawn()
		return true
	})
	ecs.Simulation.Dynamics.Range(func(d dynamic.Dynamic, _ struct{}) bool {
		if a := d.GetAnimation(); a != nil {
			a.Reset()
		}
		return true
	})
}
