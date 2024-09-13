// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"math"
	"math/rand/v2"

	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/dynamic"
	"tlyakhov/gofoom/ecs"

	"tlyakhov/gofoom/constants"
)

type PlayerController struct {
	ecs.BaseController
	*behaviors.Player
	Alive  *behaviors.Alive
	Body   *core.Body
	Mobile *core.Mobile
}

func init() {
	ecs.Types().RegisterController(&PlayerController{}, 100)
}

func (pc *PlayerController) ComponentID() ecs.ComponentID {
	return behaviors.PlayerCID
}

func (pc *PlayerController) Methods() ecs.ControllerMethod {
	return ecs.ControllerAlways
}

func (pc *PlayerController) Target(target ecs.Attachable) bool {
	pc.Player = target.(*behaviors.Player)
	if !pc.Player.IsActive() || pc.Player.Spawn {
		return false
	}
	pc.Body = core.GetBody(pc.ECS, pc.Entity)
	if pc.Body == nil || !pc.Body.IsActive() {
		return false
	}
	pc.Alive = behaviors.GetAlive(pc.ECS, pc.Entity)
	if pc.Alive == nil || !pc.Alive.IsActive() {
		return false
	}
	pc.Mobile = core.GetMobile(pc.ECS, pc.Entity)
	return pc.Mobile != nil && pc.Mobile.IsActive()
}

func (pc *PlayerController) Always() {
	uw := pc.Underwater()

	if uw {
		pc.Bob += 0.015
	} else {
		pc.Bob += pc.Mobile.Vel.Now.To2D().Length() / 64.0
	}

	for pc.Bob > math.Pi*2 {
		pc.Bob -= math.Pi * 2
	}

	// TODO: There's a bug here: this can cause a player<->floor collision that
	// has to be resolved by shoving the player upwards, making an uncrouch into
	// an unintentional jump.
	if pc.Crouching {
		pc.Body.Size.Now[1] = constants.PlayerCrouchHeight
	} else {
		pc.Body.Size.Now[1] = constants.PlayerHeight
	}

	bob := math.Sin(pc.Bob) * 1.5
	pc.CameraZ = pc.Body.Pos.Render[2] + pc.Body.Size.Render[1]*0.5 + bob - 5

	if sector := pc.Body.Sector(); sector != nil {
		fz, cz := sector.ZAt(dynamic.DynamicRender, pc.Body.Pos.Render.To2D())
		fz += constants.IntersectEpsilon
		cz -= constants.IntersectEpsilon
		if pc.CameraZ < fz {
			pc.CameraZ = fz
		}
		if pc.CameraZ > cz {
			pc.CameraZ = cz
		}
	}

	if uw {
		pc.FrameTint = concepts.Vector4{0.29, 0.58, 1, 0.35}
	} else {
		pc.FrameTint = concepts.Vector4{}
	}

	pc.Alive.Tint(&pc.FrameTint)

	/*for _, item := range pc.Inventory {
		if w := behaviors.GetWeaponInstant(pc.ECS, item.Entity); w != nil {
			pc.CurrentWeapon = item.Entity
		}
	}*/
}

func MovePlayer(db *ecs.ECS, e ecs.Entity, angle float64, direct bool) {
	p := core.GetBody(db, e)
	m := core.GetMobile(db, e)
	if p == nil || m == nil {
		return
	}
	uw := behaviors.GetUnderwater(db, p.SectorEntity) != nil
	dy, dx := math.Sincos(angle * concepts.Deg2rad)
	dy *= constants.PlayerWalkForce
	dx *= constants.PlayerWalkForce
	if direct {
		p.Pos.Now[0] += dx * 0.1 / m.Mass
		p.Pos.Now[1] += dy * 0.1 / m.Mass
	} else {
		if uw || p.OnGround {
			m.Force[0] += dx
			m.Force[1] += dy
		}
	}
}

func Respawn(db *ecs.ECS, force bool) {
	spawns := make([]*behaviors.Player, 0)
	players := make([]*behaviors.Player, 0)
	col := ecs.ColumnFor[behaviors.Player](db, behaviors.PlayerCID)
	for i := range col.Length {
		p := col.Value(i)
		if !p.Active {
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
		db.DetachAll(players[len(players)-1].Entity)
		players = players[:len(players)-1]
	}

	if len(players) > 0 || len(spawns) == 0 {
		return
	}

	spawn := spawns[rand.Int()%len(spawns)]
	copiedSpawn := db.SerializeEntity(spawn.Entity)
	var pastedEntity ecs.Entity
	for name, id := range ecs.Types().IDs {
		jsonData := copiedSpawn[name]
		if jsonData == nil {
			continue
		}
		if pastedEntity == 0 {
			pastedEntity = db.NewEntity()
		}
		jsonComponent := jsonData.(map[string]any)
		c := db.LoadComponentWithoutAttaching(id, jsonComponent)
		c = db.Attach(id, pastedEntity, c)
		if id == behaviors.PlayerCID {
			player := c.(*behaviors.Player)
			player.Spawn = false
		} else if id == ecs.NamedCID {
			named := c.(*ecs.Named)
			named.Name = "Player"
		}
	}
	db.ActAllControllersOneEntity(pastedEntity, ecs.ControllerRecalculate)
	db.ActAllControllersOneEntity(pastedEntity, ecs.ControllerAlways)
}
