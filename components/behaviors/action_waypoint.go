// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/ecs"
)

// This component represents a target position an entity should go to. Examples:
// - Animated entities and NPCs can follow this
// - Special FX can be shaped by paths
type ActionWaypoint struct {
	ActionTimed `editable:"^"`

	P concepts.Vector3 `editable:"Position"`
}

var ActionWaypointCID ecs.ComponentID

func init() {
	ActionWaypointCID = ecs.RegisterComponent(&ecs.Column[ActionWaypoint, *ActionWaypoint]{Getter: GetActionWaypoint})
}

func (*ActionWaypoint) ComponentID() ecs.ComponentID {
	return ActionWaypointCID
}
func GetActionWaypoint(u *ecs.Universe, e ecs.Entity) *ActionWaypoint {
	if asserted, ok := u.Component(e, ActionWaypointCID).(*ActionWaypoint); ok {
		return asserted
	}
	return nil
}

func (waypoint *ActionWaypoint) String() string {
	return "Waypoint: " + waypoint.P.StringHuman(2)
}

func (waypoint *ActionWaypoint) Construct(data map[string]any) {
	waypoint.ActionTimed.Construct(data)
	waypoint.P = concepts.Vector3{}

	if data == nil {
		return
	}

	if v, ok := data["P"]; ok {
		waypoint.P.Deserialize(v.(string))
	}
}

func (waypoint *ActionWaypoint) Serialize() map[string]any {
	result := waypoint.ActionTimed.Serialize()
	result["P"] = waypoint.P.Serialize()

	return result
}
