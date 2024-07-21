// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"fmt"
	"tlyakhov/gofoom/concepts"
)

type Damage struct {
	Amount   float64
	Cooldown concepts.SimVariable[float64]
}

type Alive struct {
	concepts.Attached `editable:"^"`
	Health            float64 `editable:"Health"`
	Damages           map[string]*Damage
}

var AliveComponentIndex int

func init() {
	AliveComponentIndex = concepts.DbTypes().Register(Alive{}, AliveFromDb)
}

func AliveFromDb(db *concepts.EntityComponentDB, e concepts.Entity) *Alive {
	if asserted, ok := db.Component(e, AliveComponentIndex).(*Alive); ok {
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
		d.Cooldown.Attach(a.DB.Simulation)
	}
	d.Cooldown.SetAll(cooldown)
	a.Damages[source] = &d

	return true
}

func (a *Alive) Serialize() map[string]any {
	result := a.Attached.Serialize()
	result["Health"] = a.Health

	return result
}
