// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package inventory

import (
	"tlyakhov/gofoom/ecs"

	"github.com/spf13/cast"
)

type WeaponClassProjectile struct {
	ecs.Attached `editable:"^"`

	Projectile ecs.Entity `editable:"Projectile" edit_type:"Spawner"`
}

func (w *WeaponClassProjectile) Shareable() bool { return true }

func (w *WeaponClassProjectile) String() string {
	return "WeaponClassProjectile"
}

func (w *WeaponClassProjectile) Construct(data map[string]any) {
	w.Attached.Construct(data)
	w.Projectile = 0

	if data == nil {
		return
	}

	if v, ok := data["Projectile"]; ok {
		w.Projectile, _ = ecs.ParseEntityHumanOrCanonical(cast.ToString(v))
	}
}

func (w *WeaponClassProjectile) Serialize() map[string]any {
	result := w.Attached.Serialize()

	if w.Projectile != 0 {
		result["Projectile"] = w.Projectile.Serialize()
	}

	return result
}
