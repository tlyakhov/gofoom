// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package inventory

import (
	"tlyakhov/gofoom/ecs"

	"github.com/spf13/cast"
)

type WeaponClassInstant struct {
	ecs.Attached `editable:"^"`

	Damage float64 `editable:"Damage"`
}

func (w *WeaponClassInstant) Shareable() bool { return true }

func (w *WeaponClassInstant) String() string {
	return "WeaponClassInstant"
}

func (w *WeaponClassInstant) Construct(data map[string]any) {
	w.Attached.Construct(data)
	w.Damage = 10

	if data == nil {
		return
	}

	if v, ok := data["Damage"]; ok {
		w.Damage = cast.ToFloat64(v)
	}
}

func (w *WeaponClassInstant) Serialize() map[string]any {
	result := w.Attached.Serialize()

	result["Damage"] = w.Damage

	return result
}
