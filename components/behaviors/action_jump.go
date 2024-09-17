// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"tlyakhov/gofoom/ecs"
)

type ActionJump struct {
	ActionTimed `editable:"^"`
}

var ActionJumpCID ecs.ComponentID

func init() {
	ActionJumpCID = ecs.RegisterComponent(&ecs.Column[ActionJump, *ActionJump]{Getter: GetActionJump}, "")
}

func GetActionJump(db *ecs.ECS, e ecs.Entity) *ActionJump {
	if asserted, ok := db.Component(e, ActionJumpCID).(*ActionJump); ok {
		return asserted
	}
	return nil
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
