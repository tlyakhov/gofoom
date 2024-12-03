// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/containers"
	"tlyakhov/gofoom/ecs"

	"github.com/spf13/cast"
)

type Player struct {
	ecs.Attached `editable:"^"`

	Spawn         bool
	FrameTint     concepts.Vector4
	Crouching     bool
	Bob           float64
	CameraZ       float64
	ActionPressed bool

	SelectedTarget  ecs.Entity
	HoveringTargets containers.Set[ecs.Entity]

	Notices containers.SyncUniqueQueue[string]
}

var PlayerCID ecs.ComponentID

func init() {
	PlayerCID = ecs.RegisterComponent(&ecs.Column[Player, *Player]{Getter: GetPlayer})
}

func GetPlayer(db *ecs.ECS, e ecs.Entity) *Player {
	if asserted, ok := db.Component(e, PlayerCID).(*Player); ok {
		return asserted
	}
	return nil
}

func (p *Player) Underwater() bool {
	if b := core.GetBody(p.ECS, p.Entities.First()); b != nil {
		if u := GetUnderwater(p.ECS, b.SectorEntity); u != nil {
			return true
		}
	}
	return false
}

func (p *Player) Construct(data map[string]any) {
	p.Attached.Construct(data)
	// By convention, we construct spawn points rather than active players to
	// avoid weird behaviors.
	p.Spawn = true
	p.HoveringTargets = make(containers.Set[ecs.Entity])

	if data == nil {
		return
	}

	if v, ok := data["Spawn"]; ok {
		p.Spawn = cast.ToBool(v)
	}
}

func (p *Player) Serialize() map[string]any {
	result := p.Attached.Serialize()

	result["Spawn"] = p.Spawn

	return result
}
