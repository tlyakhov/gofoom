// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package inventory

import (
	"tlyakhov/gofoom/ecs"

	"github.com/gammazero/deque"
)

//go:generate go run github.com/dmarkham/enumer -type=WeaponIntent -json
type WeaponIntent int

const (
	WeaponHeld WeaponIntent = iota
	WeaponFire
	WeaponHolstered
)

// Weapon represents the state for a weapon actually held by a player or NPC.
type Weapon struct {
	ecs.Attached `editable:"^"`

	// Internal state
	State              WeaponState
	Intent             WeaponIntent `editable:"Intent"`
	LastStateTimestamp int64        // in ms
	// TODO: We should serialize these
	Marks deque.Deque[WeaponMark]
}

func (w *Weapon) StateDuration() int64 {
	return ecs.Simulation.Timestamp - w.LastStateTimestamp
}

func (w *Weapon) StateCompleted() bool {
	if wc := GetWeaponClass(w.Entity); wc != nil {
		return float64(w.StateDuration()) >= wc.Params[w.State].Time
	}
	return false
}

func (w *Weapon) String() string {
	return "Weapon"
}

func (w *Weapon) Construct(data map[string]any) {
	w.Attached.Construct(data)
	w.Marks = deque.Deque[WeaponMark]{}

	if data == nil {
		return
	}
}

func (w *Weapon) Serialize() map[string]any {
	result := w.Attached.Serialize()

	return result
}
