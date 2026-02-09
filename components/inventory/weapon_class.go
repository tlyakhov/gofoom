// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package inventory

import (
	"tlyakhov/gofoom/ecs"

	"github.com/spf13/cast"
)

// TODO: Implement scripting
type WeaponClass struct {
	ecs.Attached `editable:"^"`

	Spread float64                             `editable:"Spread"` // In degrees
	Params [WeaponStateCount]WeaponStateParams `editable:"Params"`

	FlashMaterial ecs.Entity `editable:"Flash Material" edit_type:"Material"`
}

func (w *WeaponClass) Shareable() bool { return true }

func (w *WeaponClass) String() string {
	return "WeaponClass"
}

func (w *WeaponClass) Construct(data map[string]any) {
	w.Attached.Construct(data)
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

	if v, ok := data["Spread"]; ok {
		w.Spread = cast.ToFloat64(v)
	}

	if v, ok := data["FlashMaterial"]; ok {
		w.FlashMaterial, _ = ecs.ParseEntity(v.(string))
	}

}

func (w *WeaponClass) Serialize() map[string]any {
	result := w.Attached.Serialize()

	result["Spread"] = w.Spread

	p := make([]any, WeaponStateCount)
	for i := range w.Params {
		p[i] = w.Params[i].Serialize()
	}
	result["Params"] = p

	if w.FlashMaterial != 0 {
		result["FlashMaterial"] = w.FlashMaterial.Serialize()
	}

	return result
}
