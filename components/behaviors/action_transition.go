// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/ecs"
)

type ActionTransition struct {
	ecs.Attached `editable:"^"`

	Next ecs.EntityTable `editable:"Next" edit_type:"Action"`
}

var ActionTransitionCID ecs.ComponentID

func init() {
	ActionTransitionCID = ecs.RegisterComponent(&ecs.Column[ActionTransition, *ActionTransition]{Getter: GetActionTransition})
}

func GetActionTransition(db *ecs.ECS, e ecs.Entity) *ActionTransition {
	if asserted, ok := db.Component(e, ActionTransitionCID).(*ActionTransition); ok {
		return asserted
	}
	return nil
}

func (transition *ActionTransition) String() string {
	var s string

	for _, e := range transition.Next {
		if e == 0 {
			continue
		}
		if len(s) > 0 {
			s += ", "
		}
		s += e.NameString(transition.ECS)
	}
	return "Transition to " + s
}

func (transition *ActionTransition) Construct(data map[string]any) {
	transition.Attached.Construct(data)
	transition.Next = make(ecs.EntityTable, 0)

	if data == nil {
		return
	}

	if v, ok := data["Next"]; ok {
		if s, ok := v.(string); ok {
			transition.Next = ecs.ParseEntityCSV(s)
		} else {
			transition.Next = ecs.DeserializeEntities(v.([]any))
		}
	}
}

func (transition *ActionTransition) Serialize() map[string]any {
	result := transition.Attached.Serialize()

	if len(transition.Next) > 0 {
		result["Next"] = transition.Next.Serialize()
	}

	return result
}

func IterateActions(db *ecs.ECS, start ecs.Entity, f func(action ecs.Entity, parentPosition *concepts.Vector3)) {
	var parentP *concepts.Vector3
	visited := make(map[ecs.Entity]*concepts.Vector3, 1)
	actions := make(map[ecs.Entity]ecs.Entity)
	actions[start] = 0

	for len(actions) > 0 {
		var action ecs.Entity
		for action = range actions {
			break
		}
		// Avoid infinite loops
		if _, ok := visited[action]; ok {
			delete(actions, action)
			continue
		}

		visited[action] = nil

		if waypoint := GetActionWaypoint(db, action); waypoint != nil {
			visited[action] = &waypoint.P

			if actions[action] != 0 {
				parentP = visited[actions[action]]
			}
		}

		delete(actions, action)

		f(action, parentP)

		transition := GetActionTransition(db, action)
		if transition == nil {
			continue
		}

		for _, next := range transition.Next {
			if next == 0 {
				continue
			}

			actions[next] = action
		}
	}
}
