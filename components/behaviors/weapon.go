// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"tlyakhov/gofoom/ecs"

	"github.com/gammazero/deque"
)

// Represents state for an instant-hit weapon (think railgun rather than rocket launcher)
type Weapon struct {
	ecs.Attached `editable:"^"`

	// Internal state
	FiredTimestamp int64 // in ms
	FireNextFrame  bool  `editable:"Fire Next Frame"`
	// TODO: We should serialize these
	Marks deque.Deque[WeaponMark]
}

var WeaponCID ecs.ComponentID

func init() {
	WeaponCID = ecs.RegisterComponent(&ecs.Column[Weapon, *Weapon]{Getter: GetWeapon})
}

func GetWeapon(u *ecs.Universe, e ecs.Entity) *Weapon {
	if asserted, ok := u.Component(e, WeaponCID).(*Weapon); ok {
		return asserted
	}
	return nil
}

func (w *Weapon) CoolingDown() bool {
	wc := GetWeaponClass(w.Universe, w.Entity)
	return wc != nil && w.Universe.Timestamp-w.FiredTimestamp < int64(wc.Cooldown)
}

func (w *Weapon) Flashing() bool {
	wc := GetWeaponClass(w.Universe, w.Entity)
	return wc != nil && w.Universe.Timestamp-w.FiredTimestamp < int64(wc.FlashTime)
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
