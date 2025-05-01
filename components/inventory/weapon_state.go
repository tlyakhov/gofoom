// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package inventory

import (
	"tlyakhov/gofoom/ecs"

	"github.com/spf13/cast"
)

//go:generate go run github.com/dmarkham/enumer -type=WeaponState -json
type WeaponState int

const (
	WeaponIdle WeaponState = iota
	WeaponUnholstering
	WeaponFiring
	WeaponCooling
	WeaponReloading
	WeaponHolstering
	WeaponStateCount
)

type WeaponStateParams struct {
	Time     float64    `editable:"Time"` // In ms
	Material ecs.Entity `editable:"Material" edit_type:"Material"`
}

func (w *WeaponStateParams) Construct(data map[string]any) {
	w.Time = 100

	if data == nil {
		return
	}

	if v, ok := data["Time"]; ok {
		w.Time = cast.ToFloat64(v)
	}

	if v, ok := data["Material"]; ok {
		w.Material, _ = ecs.ParseEntity(v.(string))
	}
}

func (w *WeaponStateParams) Serialize(u *ecs.Universe) map[string]any {
	result := make(map[string]any)

	result["Time"] = w.Time

	if w.Material != 0 {
		result["Material"] = w.Material.Serialize(u)
	}

	return result
}
