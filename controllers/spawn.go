// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"math/rand"
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/dynamic"
	"tlyakhov/gofoom/ecs"
)

func Respawn(db *ecs.ECS, force bool) {
	spawns := make([]*behaviors.Player, 0)
	players := make([]*behaviors.Player, 0)
	col := ecs.ColumnFor[behaviors.Player](db, behaviors.PlayerCID)
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
		db.Delete(players[len(players)-1].Entity)
		players = players[:len(players)-1]
	}

	if len(players) > 0 || len(spawns) == 0 {
		return
	}

	spawn := spawns[rand.Int()%len(spawns)]
	copiedSpawn := db.SerializeEntity(spawn.Entity)
	var pastedEntity ecs.Entity
	for name, cid := range ecs.Types().IDs {
		jsonData := copiedSpawn[name]
		if jsonData == nil {
			continue
		}
		if pastedEntity == 0 {
			pastedEntity = db.NewEntity()
		}
		jsonComponent := jsonData.(map[string]any)
		c := db.LoadComponentWithoutAttaching(cid, jsonComponent)
		c = db.Attach(cid, pastedEntity, c)
		if cid == behaviors.PlayerCID {
			player := c.(*behaviors.Player)
			player.Spawn = false
		} else if cid == ecs.NamedCID {
			named := c.(*ecs.Named)
			named.Name = "Player"
		}
	}
	db.ActAllControllersOneEntity(pastedEntity, ecs.ControllerRecalculate)
	db.ActAllControllersOneEntity(pastedEntity, ecs.ControllerAlways)
}

func ResetAllSpawnables(ecs *ecs.ECS) {
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
