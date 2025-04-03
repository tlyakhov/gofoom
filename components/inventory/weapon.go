// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package inventory

import (
	"tlyakhov/gofoom/ecs"

	"github.com/gammazero/deque"
)

// Weapon represents the state for a weapon actually held by a player or NPC.
// Here's how all these components are connected:
//   - A Player or NPC has an Carrier component attached, which enables
//     that character to have an array of InventorySlots.
//   - InventorySlot can have a WeaponClass component attached to it, which would
//     indicate the traits of the weapon the character is holding.
//   - WeaponClass describes the traits of a given type of weapon. The WeaponClass
//     could be unique for that particular NPC/player, or it could be a common
//     component.
//   - Weapon components hold the state of a particular player/NPC's weapon. They
//     should be created automatically by the game as necessary, see InventorySlot controller.
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

func (x *Weapon) ComponentID() ecs.ComponentID {
	return WeaponCID
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
