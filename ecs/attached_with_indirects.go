// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ecs

// See ecs.act for how this is applied
type AttachedWithIndirects struct {
	Attached `editable:"^"`

	ApplyToEntities EntityTable `editable:"Apply To Entities"`
}

func (ai *AttachedWithIndirects) Indirects() *EntityTable {
	return &ai.ApplyToEntities
}

func (ai *AttachedWithIndirects) Construct(data map[string]any) {
	ai.Attached.Construct(data)
	ai.ApplyToEntities = nil

	if data == nil {
		return
	}

	if v, ok := data["ApplyToEntities"]; ok {
		ai.ApplyToEntities = ParseEntityTable(v, true)
	}
}

func (ai *AttachedWithIndirects) Serialize() map[string]any {
	result := ai.Attached.Serialize()

	if len(ai.ApplyToEntities) > 0 {
		result["ApplyToEntities"] = ai.ApplyToEntities.Serialize()
	}
	return result
}
