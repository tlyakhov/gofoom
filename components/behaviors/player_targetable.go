// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/ecs"
)

type PlayerTargetable struct {
	ecs.Attached `editable:"^"`
	Frob         core.Script `editable:"Frob"`
	Selected     core.Script `editable:"Selected"`
	UnSelected   core.Script `editable:"UnSelected"`
	Message      string      `editable:"Message"`
}

var PlayerTargetableCID ecs.ComponentID

func init() {
	PlayerTargetableCID = ecs.RegisterComponent(&ecs.Column[PlayerTargetable, *PlayerTargetable]{Getter: GetPlayerTargetable})
}

func GetPlayerTargetable(db *ecs.ECS, e ecs.Entity) *PlayerTargetable {
	if asserted, ok := db.Component(e, PlayerTargetableCID).(*PlayerTargetable); ok {
		return asserted
	}
	return nil
}

func (pt *PlayerTargetable) String() string {
	return "PlayerTargetable"
}

func (pt *PlayerTargetable) OnDetach() {
	if pt.ECS != nil {
		pt.Frob.ECS = nil
		pt.Selected.ECS = nil
		pt.UnSelected.ECS = nil
	}
	pt.Attached.OnDetach()
}

func (pt *PlayerTargetable) AttachECS(db *ecs.ECS) {
	if pt.ECS != db {
		pt.OnDetach()
	}
	pt.Attached.AttachECS(db)
	pt.Frob.AttachECS(db)
	pt.Selected.AttachECS(db)
	pt.UnSelected.AttachECS(db)
}

func (pt *PlayerTargetable) Construct(data map[string]any) {
	pt.Attached.Construct(data)

	if data == nil {
		return
	}

	if v, ok := data["Frob"]; ok {
		pt.Frob.Construct(v.(map[string]any))
	}
	if v, ok := data["Selected"]; ok {
		pt.Selected.Construct(v.(map[string]any))
	}
	if v, ok := data["UnSelected"]; ok {
		pt.UnSelected.Construct(v.(map[string]any))
	}
	if v, ok := data["Message"]; ok {
		pt.Message = v.(string)
	}
}

func (pt *PlayerTargetable) Serialize() map[string]any {
	result := pt.Attached.Serialize()

	if !pt.Frob.IsEmpty() {
		result["Frob"] = pt.Frob.Serialize()
	}
	if !pt.Selected.IsEmpty() {
		result["Selected"] = pt.Selected.Serialize()
	}
	if !pt.UnSelected.IsEmpty() {
		result["UnSelected"] = pt.UnSelected.Serialize()
	}
	if len(pt.Message) > 0 {
		result["Message"] = pt.Message
	}
	return result
}
