// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"

	"github.com/gammazero/deque"
)

// TODO: Implement attributes like cooldowns, spread, DPS, scripting
type WeaponInstant struct {
	concepts.Attached `editable:"^"`

	// Bullets make marks on walls/internal segments
	MarkMaterial concepts.Entity `editable:"Mark Material" edit_type:"Material"`
	MarkSize     float64         `editable:"Mark Size"`

	// Internal state
	FireNextFrame bool `editable:"Fire Next Frame"`
	// TODO: We should serialize these
	Marks deque.Deque[WeaponMark]
}

type WeaponMark struct {
	*materials.ShaderStage
	*materials.Surface
}

var WeaponInstantComponentIndex int

func init() {
	WeaponInstantComponentIndex = concepts.DbTypes().Register(WeaponInstant{}, WeaponInstantFromDb)
}

func WeaponInstantFromDb(db *concepts.EntityComponentDB, e concepts.Entity) *WeaponInstant {
	if asserted, ok := db.Component(e, WeaponInstantComponentIndex).(*WeaponInstant); ok {
		return asserted
	}
	return nil
}

func (w *WeaponInstant) String() string {
	return "WeaponInstant"
}

func (w *WeaponInstant) Construct(data map[string]any) {
	w.Attached.Construct(data)
	w.MarkSize = 5
	w.Marks = deque.Deque[WeaponMark]{}

	if data == nil {
		return
	}

	if v, ok := data["MarkMaterial"]; ok {
		w.MarkMaterial, _ = concepts.ParseEntity(v.(string))
	}

	if v, ok := data["MarkSize"]; ok {
		w.MarkSize = v.(float64)
	}
}

func (w *WeaponInstant) Serialize() map[string]any {
	result := w.Attached.Serialize()

	if w.MarkMaterial != 0 {
		result["MarkMaterial"] = w.MarkMaterial.Format()
	}

	if w.MarkSize != 5 {
		result["MarkSize"] = w.MarkSize
	}

	return result
}
