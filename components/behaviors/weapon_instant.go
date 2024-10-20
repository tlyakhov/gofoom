// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"tlyakhov/gofoom/ecs"

	"github.com/gammazero/deque"
)

type WeaponInstant struct {
	ecs.Attached `editable:"^"`

	Class ecs.Entity `editable:"Class" edit_type:"Weapon"`

	// Internal state
	FireNextFrame bool `editable:"Fire Next Frame"`
	// TODO: We should serialize these
	Marks deque.Deque[WeaponMark]
}

var WeaponInstantCID ecs.ComponentID

func init() {
	WeaponInstantCID = ecs.RegisterComponent(&ecs.Column[WeaponInstant, *WeaponInstant]{Getter: GetWeaponInstant})
}

func GetWeaponInstant(db *ecs.ECS, e ecs.Entity) *WeaponInstant {
	if asserted, ok := db.Component(e, WeaponInstantCID).(*WeaponInstant); ok {
		return asserted
	}
	return nil
}

func (w *WeaponInstant) String() string {
	return "WeaponInstant"
}

func (w *WeaponInstant) Construct(data map[string]any) {
	w.Attached.Construct(data)
	w.Marks = deque.Deque[WeaponMark]{}

	if data == nil {
		return
	}

	if v, ok := data["Class"]; ok {
		w.Class, _ = ecs.ParseEntity(v.(string))
	}
}

func (w *WeaponInstant) Serialize() map[string]any {
	result := w.Attached.Serialize()

	if w.Class != 0 {
		result["Class"] = w.Class.String()
	}

	return result
}
