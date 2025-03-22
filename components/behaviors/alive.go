// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"fmt"
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
	Health       float64 `editable:"Health"`
	Damages      map[string]*Damage
}

var AliveCID ecs.ComponentID

func init() {
	AliveCID = ecs.RegisterComponent(&ecs.Column[Alive, *Alive]{Getter: GetAlive})
}

func GetAlive(u *ecs.Universe, e ecs.Entity) *Alive {
	if asserted, ok := u.Component(e, AliveCID).(*Alive); ok {
		return asserted
	}
	return nil
}

func (a *Alive) MultiAttachable() bool {
	return true
}

func (a *Alive) String() string {
	return fmt.Sprintf("Alive: %.2f", a.Health)
}

func (a *Alive) Construct(data map[string]any) {
	a.Attached.Construct(data)

	a.Health = 100
	a.Damages = make(map[string]*Damage)

	if data == nil {
		return
	}

	if v, ok := data["Health"]; ok {
		a.Health = cast.ToFloat64(v)
	}
}

func (a *Alive) Hurt(source string, amount, cooldown float64) bool {
	if _, ok := a.Damages[source]; ok {
		return false
	}
	d := Damage{Amount: amount}
	if !d.Cooldown.Attached {
		d.Cooldown.Attach(a.Universe.Simulation)
	}
	d.Cooldown.SetAll(cooldown)
	a.Damages[source] = &d

	return true
}

var damageTintColor = concepts.Vector4{1, 0, 0, 1}

func (alive *Alive) Tint(color *concepts.Vector4) {
	allCooldowns := 0.0
	maxCooldown := 0.0
	for _, d := range alive.Damages {
		allCooldowns += *d.Cooldown.Render
		maxCooldown += d.Cooldown.Spawn
	}

	if allCooldowns > 0 && maxCooldown > 0 {
		a := allCooldowns * 0.6 / maxCooldown
		concepts.BlendColors(color, &damageTintColor, a)
	}
}

func (a *Alive) Serialize() map[string]any {
	result := a.Attached.Serialize()
	result["Health"] = a.Health

	return result
}
