// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"fmt"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/dynamic"
	"tlyakhov/gofoom/ecs"

	"github.com/spf13/cast"
)

type Damage struct {
	Amount float64
	// This is in frames currently, should be in seconds or ms
	Cooldown dynamic.DynamicValue[float64]
}

type Alive struct {
	ecs.Attached `editable:"^"`
	Health       dynamic.DynamicValue[float64] `editable:"Health"`
	Faction      string                        `editable:"Faction"`

	Damages map[string]*Damage

	Die  core.Script `editable:"Die"`
	Live core.Script `editable:"Live"`
}

func (a *Alive) Shareable() bool {
	return true
}

func (a *Alive) String() string {
	return fmt.Sprintf("Alive: %.2f (%.2f)", a.Health.Now, a.Health.Spawn)
}

func (a *Alive) OnDelete() {
	defer a.Attached.OnDelete()
	if a.IsAttached() {
		a.Health.Detach(ecs.Simulation)
	}
}

func (a *Alive) OnAttach() {
	a.Attached.OnAttach()
	a.Health.Attach(ecs.Simulation)
}

func (a *Alive) Construct(data map[string]any) {
	a.Attached.Construct(data)

	a.Health.SetAll(100)
	a.Damages = make(map[string]*Damage)
	a.Faction = ""

	if data == nil {
		a.Die.Construct(nil)
		a.Live.Construct(nil)
		return
	}

	if v, ok := data["Health"]; ok {
		a.Health.Construct(v)
	}
	if v, ok := data["Faction"]; ok {
		a.Faction = cast.ToString(v)
	}
	if v, ok := data["Die"]; ok {
		a.Die.Construct(v.(map[string]any))
	} else {
		a.Die.Construct(nil)
	}
	if v, ok := data["Live"]; ok {
		a.Live.Construct(v.(map[string]any))
	} else {
		a.Live.Construct(nil)
	}
}

func (a *Alive) Hurt(source string, amount, cooldown float64) bool {
	if _, ok := a.Damages[source]; ok {
		return false
	}
	d := Damage{Amount: amount}
	if !d.Cooldown.Attached {
		d.Cooldown.Attach(ecs.Simulation)
	}
	d.Cooldown.SetAll(cooldown)
	a.Damages[source] = &d

	return true
}

func (a *Alive) Tint(color, damageTintColor *concepts.Vector4) {
	allCooldowns := 0.0
	maxCooldown := 0.0
	for _, d := range a.Damages {
		allCooldowns += d.Cooldown.Render
		maxCooldown += d.Cooldown.Spawn
	}

	if allCooldowns > 0 && maxCooldown > 0 {
		a := allCooldowns * 0.6 / maxCooldown
		concepts.BlendColors(color, damageTintColor, a)
	}
}

func (a *Alive) Serialize() map[string]any {
	result := a.Attached.Serialize()
	result["Health"] = a.Health.Serialize()
	if a.Faction != "" {
		result["Faction"] = a.Faction
	}
	if !a.Die.IsEmpty() {
		result["Die"] = a.Die.Serialize()
	}
	if !a.Live.IsEmpty() {
		result["Live"] = a.Live.Serialize()
	}
	return result
}
