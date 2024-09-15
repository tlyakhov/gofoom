// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"tlyakhov/gofoom/containers"
	"tlyakhov/gofoom/ecs"
)

type ActionFire struct {
	ecs.Attached `editable:"^"`

	Fired containers.Set[ecs.Entity]
}

var ActionFireCID ecs.ComponentID

func init() {
	ActionFireCID = ecs.RegisterComponent(&ecs.Column[ActionFire, *ActionFire]{Getter: GetActionFire}, "")
}

func GetActionFire(db *ecs.ECS, e ecs.Entity) *ActionFire {
	if asserted, ok := db.Component(e, ActionFireCID).(*ActionFire); ok {
		return asserted
	}
	return nil
}

func (fire *ActionFire) String() string {
	return "Fire"
}

func (fire *ActionFire) Construct(data map[string]any) {
	fire.Attached.Construct(data)

	fire.Fired = make(containers.Set[ecs.Entity])

	if data == nil {
		return
	}
}

func (fire *ActionFire) Serialize() map[string]any {
	result := fire.Attached.Serialize()

	return result
}
