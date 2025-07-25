// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"tlyakhov/gofoom/ecs"
)

type ActionFire struct {
	ActionTimed `editable:"^"`
}

var ActionFireCID ecs.ComponentID

func init() {
	ActionFireCID = ecs.RegisterComponent(&ecs.Arena[ActionFire, *ActionFire]{Getter: GetActionFire})
}

func (*ActionFire) ComponentID() ecs.ComponentID {
	return ActionFireCID
}
func GetActionFire(e ecs.Entity) *ActionFire {
	if asserted, ok := ecs.Component(e, ActionFireCID).(*ActionFire); ok {
		return asserted
	}
	return nil
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
