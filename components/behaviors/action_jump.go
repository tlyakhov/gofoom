// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

type ActionJump struct {
	ActionTimed `editable:"^"`
}

func (jump *ActionJump) String() string {
	return "Jump"
}

func (jump *ActionJump) Construct(data map[string]any) {
	jump.ActionTimed.Construct(data)

	if data == nil {
		return
	}
}

func (jump *ActionJump) Serialize() map[string]any {
	result := jump.ActionTimed.Serialize()

	return result
}
