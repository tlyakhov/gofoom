// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"tlyakhov/gofoom/concepts"

	"github.com/spf13/cast"
)

// This component represents a target position an entity should go to. Examples:
// - Animated entities and NPCs can follow this
// - Special FX can be shaped by paths
type ActionWaypoint struct {
	ActionTimed `editable:"^"`

	P             concepts.Vector3 `editable:"Position"`
	UsePathFinder bool             `editable:"Use PathFinder"`
}

func (waypoint *ActionWaypoint) String() string {
	return "Waypoint: " + waypoint.P.StringHuman(2)
}

func (waypoint *ActionWaypoint) Construct(data map[string]any) {
	waypoint.ActionTimed.Construct(data)
	waypoint.P = concepts.Vector3{}
	waypoint.UsePathFinder = true

	if data == nil {
		return
	}

	if v, ok := data["P"]; ok {
		waypoint.P.Deserialize(v.(string))
	}
	if v, ok := data["UsePathFinder"]; ok {
		waypoint.UsePathFinder = cast.ToBool(v)
	}
}

func (waypoint *ActionWaypoint) Serialize() map[string]any {
	result := waypoint.ActionTimed.Serialize()
	result["P"] = waypoint.P.Serialize()
	if !waypoint.UsePathFinder {
		result["UsePathFinder"] = waypoint.UsePathFinder
	}
	return result
}
