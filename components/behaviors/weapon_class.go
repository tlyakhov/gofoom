// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/ecs"

	"github.com/spf13/cast"
)

// TODO: Implement scripting
// TODO: Muzzle flash
type WeaponClass struct {
	ecs.Attached `editable:"^"`

	InstantHit bool `editable:"InstantHit"`

	Damage   float64 `editable:"Damage"`
	Spread   float64 `editable:"Spread"`   // In degrees
	Cooldown float64 `editable:"Cooldown"` // In ms

	FlashTime     float64    `editable:"Flash Time"` // In ms
	FlashMaterial ecs.Entity `editable:"Flash Material" edit_type:"Material"`

	// Projectiles make marks on walls/internal segments
	MarkMaterial ecs.Entity `editable:"Mark Material" edit_type:"Material"`
	MarkSize     float64    `editable:"Mark Size"`
}

type WeaponMark struct {
	*materials.ShaderStage
	*materials.Surface
}

var WeaponClassCID ecs.ComponentID

func init() {
	WeaponClassCID = ecs.RegisterComponent(&ecs.Column[WeaponClass, *WeaponClass]{Getter: GetWeaponClass})
}

func GetWeaponClass(db *ecs.ECS, e ecs.Entity) *WeaponClass {
	if asserted, ok := db.Component(e, WeaponClassCID).(*WeaponClass); ok {
		return asserted
	}
	return nil
}

func (w *WeaponClass) String() string {
	return "WeaponClass"
}

func (w *WeaponClass) Construct(data map[string]any) {
	w.Attached.Construct(data)
	w.MarkSize = 5
	w.Damage = 10
	w.Spread = 1
	w.Cooldown = 100
	w.FlashTime = 100

	if data == nil {
		return
	}

	if v, ok := data["Damage"]; ok {
		w.Damage = cast.ToFloat64(v)
	}

	if v, ok := data["Spread"]; ok {
		w.Spread = cast.ToFloat64(v)
	}

	if v, ok := data["Cooldown"]; ok {
		w.Cooldown = cast.ToFloat64(v)
	}

	if v, ok := data["FlashTime"]; ok {
		w.FlashTime = cast.ToFloat64(v)
	}

	if v, ok := data["MarkMaterial"]; ok {
		w.MarkMaterial, _ = ecs.ParseEntity(v.(string))
	}

	if v, ok := data["FlashMaterial"]; ok {
		w.FlashMaterial, _ = ecs.ParseEntity(v.(string))
	}

	if v, ok := data["MarkSize"]; ok {
		w.MarkSize = cast.ToFloat64(v)
	}
}

func (w *WeaponClass) Serialize() map[string]any {
	result := w.Attached.Serialize()

	result["Damage"] = w.Damage
	result["Spread"] = w.Spread
	result["Cooldown"] = w.Cooldown
	result["FlashTime"] = w.FlashTime

	if w.MarkMaterial != 0 {
		result["MarkMaterial"] = w.MarkMaterial.Serialize(w.ECS)
	}

	if w.FlashMaterial != 0 {
		result["FlashMaterial"] = w.FlashMaterial.Serialize(w.ECS)
	}

	if w.MarkSize != 5 {
		result["MarkSize"] = w.MarkSize
	}

	return result
}
