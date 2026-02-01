// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package character

import (
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/ecs"

	"github.com/spf13/cast"
)

// TODO: What states do we actually need?
type NpcState int

//go:generate go run github.com/dmarkham/enumer -type=NpcState -json
const (
	NpcStateIdle NpcState = iota
	NpcStateSawTarget
	NpcStatePursuit
	NpcStateLostTarget
	NpcStateSearching
	NpcStateDead
	NpcStateCount
)

type Npc struct {
	ecs.Attached `editable:"^"`

	NextState NpcState `editable:"Next State"`

	BarksTargetSeen ecs.EntityTable `editable:"Barks (target seen)"`
	BarksTargetLost ecs.EntityTable `editable:"Barks (target lost)"`
	BarksIdle       ecs.EntityTable `editable:"Barks (idle)"`
	BarksFiring     ecs.EntityTable `editable:"Barks (firing)"`
	BarksHurtLow    ecs.EntityTable `editable:"Barks (hurt low)"`
	BarksHurtMed    ecs.EntityTable `editable:"Barks (hurt med)"`
	BarksHurtHigh   ecs.EntityTable `editable:"Barks (hurt high)"`
	BarksDying      ecs.EntityTable `editable:"Barks (dying)"`

	HudAlarmed ecs.Entity `editable:"HUD (alarmed)" edit_type:"Material"`

	// Internal state
	NextIdleBark   int64
	LastFiringBark int64
	State          NpcState
}

func (npc *Npc) String() string {
	return "NPC"
}

func (npc *Npc) Underwater() bool {
	if b := core.GetBody(npc.Entities.First()); b != nil {
		if u := behaviors.GetUnderwater(b.SectorEntity); u != nil {
			return true
		}
	}
	return false
}

func (npc *Npc) Construct(data map[string]any) {
	npc.Attached.Construct(data)
	npc.BarksTargetSeen = ecs.EntityTable{}
	npc.BarksTargetLost = ecs.EntityTable{}
	npc.BarksIdle = ecs.EntityTable{}
	npc.BarksFiring = ecs.EntityTable{}
	npc.BarksHurtLow = ecs.EntityTable{}
	npc.BarksHurtMed = ecs.EntityTable{}
	npc.BarksHurtHigh = ecs.EntityTable{}
	npc.BarksDying = ecs.EntityTable{}
	npc.State = NpcStateIdle
	npc.NextState = NpcStateIdle
	npc.HudAlarmed = 0

	if data == nil {
		return
	}

	if v, ok := data["BarksTargetSeen"]; ok {
		npc.BarksTargetSeen = ecs.ParseEntityTable(v, true)
	}
	if v, ok := data["BarksTargetLost"]; ok {
		npc.BarksTargetLost = ecs.ParseEntityTable(v, true)
	}
	if v, ok := data["BarksIdle"]; ok {
		npc.BarksIdle = ecs.ParseEntityTable(v, true)
	}
	if v, ok := data["BarksFiring"]; ok {
		npc.BarksFiring = ecs.ParseEntityTable(v, true)
	}
	if v, ok := data["BarksHurtLow"]; ok {
		npc.BarksHurtLow = ecs.ParseEntityTable(v, true)
	}
	if v, ok := data["BarksHurtMed"]; ok {
		npc.BarksHurtMed = ecs.ParseEntityTable(v, true)
	}
	if v, ok := data["BarksHurtHigh"]; ok {
		npc.BarksHurtHigh = ecs.ParseEntityTable(v, true)
	}
	if v, ok := data["BarksDying"]; ok {
		npc.BarksDying = ecs.ParseEntityTable(v, true)
	}
	if v, ok := data["HudAlarmed"]; ok {
		npc.HudAlarmed, _ = ecs.ParseEntityHumanOrCanonical(cast.ToString(v))
	}
}

func (npc *Npc) Serialize() map[string]any {
	result := npc.Attached.Serialize()

	if len(npc.BarksTargetSeen) > 0 {
		result["BarksTargetSeen"] = npc.BarksTargetSeen.Serialize()
	}
	if len(npc.BarksTargetLost) > 0 {
		result["BarksTargetLost"] = npc.BarksTargetLost.Serialize()
	}
	if len(npc.BarksIdle) > 0 {
		result["BarksIdle"] = npc.BarksIdle.Serialize()
	}
	if len(npc.BarksFiring) > 0 {
		result["BarksFiring"] = npc.BarksFiring.Serialize()
	}
	if len(npc.BarksHurtLow) > 0 {
		result["BarksHurtLow"] = npc.BarksHurtLow.Serialize()
	}
	if len(npc.BarksHurtMed) > 0 {
		result["BarksHurtMed"] = npc.BarksHurtMed.Serialize()
	}
	if len(npc.BarksHurtHigh) > 0 {
		result["BarksHurtHigh"] = npc.BarksHurtHigh.Serialize()
	}
	if len(npc.BarksDying) > 0 {
		result["BarksDying"] = npc.BarksDying.Serialize()
	}
	if npc.HudAlarmed != 0 {
		result["HudAlarmed"] = npc.HudAlarmed.Serialize()
	}
	return result
}
