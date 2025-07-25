// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"tlyakhov/gofoom/ecs"
)

type ActorState struct {
	ecs.Attached

	Action         ecs.Entity `editable:"Current Action" edit_type:"Action"`
	LastTransition int64
}

var ActorStateCID ecs.ComponentID

func init() {
	ActorStateCID = ecs.RegisterComponent(&ecs.Arena[ActorState, *ActorState]{Getter: GetActorState})
}

func (*ActorState) ComponentID() ecs.ComponentID {
	return ActorStateCID
}
func GetActorState(e ecs.Entity) *ActorState {
	if asserted, ok := ecs.Component(e, ActorStateCID).(*ActorState); ok {
		return asserted
	}
	return nil
}

func (a *ActorState) String() string {
	return "ActorState"
}

func (a *ActorState) Construct(data map[string]any) {
	a.Attached.Construct(data)
	a.Action = 0
	a.LastTransition = 0

	if data == nil {
		return
	}

	if v, ok := data["Action"]; ok {
		a.Action, _ = ecs.ParseEntity(v.(string))
	}
}

func (a *ActorState) Serialize() map[string]any {
	result := a.Attached.Serialize()

	result["Action"] = a.Action.Serialize()

	return result
}
