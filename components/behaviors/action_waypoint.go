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
	ecs.Attached `editable:"^"`

	P concepts.Vector3 `editable:"Position"`
}

var ActionWaypointCID ecs.ComponentID

func init() {
	ActionWaypointCID = ecs.RegisterComponent(&ecs.Column[ActionWaypoint, *ActionWaypoint]{Getter: GetActionWaypoint}, "")
}

func GetActionWaypoint(db *ecs.ECS, e ecs.Entity) *ActionWaypoint {
	if asserted, ok := db.Component(e, ActionWaypointCID).(*ActionWaypoint); ok {
		return asserted
	}
	return nil
}

func (waypoint *ActionWaypoint) String() string {
	return "Waypoint: " + waypoint.P.StringHuman()
}

func (waypoint *ActionWaypoint) Construct(data map[string]any) {
	waypoint.Attached.Construct(data)
	waypoint.P = concepts.Vector3{}

	if data == nil {
		return
	}

	waypoint.P.Deserialize(data)
}

func (waypoint *ActionWaypoint) Serialize() map[string]any {
	result := waypoint.Attached.Serialize()
	result["X"] = waypoint.P[0]
	result["Y"] = waypoint.P[1]
	result["Z"] = waypoint.P[2]

	return result
}
