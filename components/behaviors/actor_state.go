// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/ecs"
	"tlyakhov/gofoom/pathfinding"
)

type ActorState struct {
	ecs.Attached

	Action            ecs.Entity `editable:"Current Action" edit_type:"Action"`
	LastTransition    int64
	Finder            *pathfinding.Finder
	Path              []concepts.Vector3
	LastPathGenerated int64
}

func (a *ActorState) String() string {
	return "ActorState"
}

func (a *ActorState) Construct(data map[string]any) {
	a.Attached.Construct(data)
	a.Action = 0
	a.LastTransition = 0
	a.Path = nil
	a.Finder = nil

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
