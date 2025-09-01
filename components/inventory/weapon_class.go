// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package inventory

import (
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/ecs"

	"github.com/spf13/cast"
)

// TODO: Implement scripting
type WeaponClass struct {
	ecs.Attached `editable:"^"`

	InstantHit bool `editable:"InstantHit"`

	Damage float64                             `editable:"Damage"`
	Spread float64                             `editable:"Spread"` // In degrees
	Params [WeaponStateCount]WeaponStateParams `editable:"Params"`

	FlashMaterial ecs.Entity `editable:"Flash Material" edit_type:"Material"`

	// Projectiles make marks on walls/internal segments
	MarkMaterial ecs.Entity `editable:"Mark Material" edit_type:"Material"`
	MarkSize     float64    `editable:"Mark Size"`
}

type WeaponMark struct {
	*materials.ShaderStage
	*materials.Surface
}

func (w *WeaponClass) MultiAttachable() bool { return true }

func (w *WeaponClass) String() string {
	return "WeaponClass"
}

func (w *WeaponClass) Construct(data map[string]any) {
	w.Attached.Construct(data)
	w.MarkSize = 5
	w.Damage = 10
	w.Spread = 1

	for i := range WeaponStateCount {
		w.Params[i].Construct(nil)
	}

	if data == nil {
		return
	}

	if v, ok := data["Params"]; ok {
		arr := v.([]any)
		for i := range min(int(WeaponStateCount), len(arr)) {
			w.Params[i].Construct(arr[i].(map[string]any))
		}
	}

	if v, ok := data["Damage"]; ok {
		w.Damage = cast.ToFloat64(v)
	}

	if v, ok := data["Spread"]; ok {
		w.Spread = cast.ToFloat64(v)
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

	p := make([]map[string]any, WeaponStateCount)
	for i := range w.Params {
		p[i] = w.Params[i].Serialize()
	}
	result["Params"] = p

	if w.MarkMaterial != 0 {
		result["MarkMaterial"] = w.MarkMaterial.Serialize()
	}

	if w.FlashMaterial != 0 {
		result["FlashMaterial"] = w.FlashMaterial.Serialize()
	}

	if w.MarkSize != 5 {
		result["MarkSize"] = w.MarkSize
	}

	return result
}
