// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package character

import (
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/containers"
	"tlyakhov/gofoom/ecs"
)

type Player struct {
	ecs.Attached `editable:"^"`

	// TODO: should we serialize any of this?
	FrameTint     concepts.Vector4
	Crouching     bool
	Bob           float64
	CameraZ       float64
	Pitch         float64
	ActionPressed bool

	SelectedTarget  ecs.Entity
	HoveringTargets containers.Set[ecs.Entity]

	Notices       containers.SyncUniqueQueue[string]
	FrobReadyTime int64
}

func (p *Player) String() string {
	return "Player"
}

func (p *Player) Underwater() bool {
	if b := core.GetBody(p.Entities.First()); b != nil {
		if u := behaviors.GetUnderwater(b.SectorEntity); u != nil {
			return true
		}
	}
	return false
}

func (p *Player) Construct(data map[string]any) {
	p.Attached.Construct(data)
	p.HoveringTargets = make(containers.Set[ecs.Entity])

	if data == nil {
		return
	}
}

func (p *Player) Serialize() map[string]any {
	result := p.Attached.Serialize()

	return result
}
