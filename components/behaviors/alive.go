// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"fmt"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/dynamic"
	"tlyakhov/gofoom/ecs"
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
	AliveCID = ecs.RegisterComponent(&ecs.Column[Alive, *Alive]{Getter: GetAlive}, "")
}

func GetAlive(db *ecs.ECS, e ecs.Entity) *Alive {
	if asserted, ok := db.Component(e, AliveCID).(*Alive); ok {
		return asserted
	}
	return nil
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
		a.Health = v.(float64)
	}
}

func (a *Alive) Hurt(source string, amount, cooldown float64) bool {
	if _, ok := a.Damages[source]; ok {
		return false
	}
	d := Damage{Amount: amount}
	if !d.Cooldown.Attached {
		d.Cooldown.Attach(a.ECS.Simulation)
	}
	d.Cooldown.SetAll(cooldown)
	a.Damages[source] = &d

	return true
}

func (alive *Alive) Tint(color *concepts.Vector4) {
	allCooldowns := 0.0
	maxCooldown := 0.0
	for _, d := range alive.Damages {
		allCooldowns += *d.Cooldown.Render
		maxCooldown += d.Cooldown.Original
	}

	if allCooldowns > 0 && maxCooldown > 0 {
		a := allCooldowns * 0.6 / maxCooldown
		color.AddPreMulColorSelf(&concepts.Vector4{1, 0, 0, a})
	}
}

func (a *Alive) Serialize() map[string]any {
	result := a.Attached.Serialize()
	result["Health"] = a.Health

	return result
}
