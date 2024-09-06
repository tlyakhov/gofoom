// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"tlyakhov/gofoom/ecs"
)

type ActionTransition struct {
	ecs.Attached `editable:"^"`

	Next ecs.Entity `editable:"Next" edit_type:"Action"`
}

var ActionTransitionCID ecs.ComponentID

func init() {
	ActionTransitionCID = ecs.RegisterComponent(&ecs.Column[ActionTransition, *ActionTransition]{Getter: GetActionTransition}, "")
}

func GetActionTransition(db *ecs.ECS, e ecs.Entity) *ActionTransition {
	if asserted, ok := db.Component(e, ActionTransitionCID).(*ActionTransition); ok {
		return asserted
	}
	return nil
}

func (transition *ActionTransition) String() string {
	return "Transition to " + transition.Next.NameString(transition.ECS)
}

func (transition *ActionTransition) Construct(data map[string]any) {
	transition.Attached.Construct(data)

	if data == nil {
		return
	}

	if v, ok := data["Next"]; ok {
		transition.Next, _ = ecs.ParseEntity(v.(string))
	}
}

func (transition *ActionTransition) Serialize() map[string]any {
	result := transition.Attached.Serialize()

	if transition.Next != 0 {
		result["Next"] = transition.Next.String()
	}

	return result
}
