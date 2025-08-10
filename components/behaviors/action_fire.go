// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

type ActionFire struct {
	ActionTimed `editable:"^"`
}

func (fire *ActionFire) String() string {
	return "Fire"
}

func (fire *ActionFire) Construct(data map[string]any) {
	fire.ActionTimed.Construct(data)

	if data == nil {
		return
	}
}

func (fire *ActionFire) Serialize() map[string]any {
	result := fire.ActionTimed.Serialize()

	return result
}
