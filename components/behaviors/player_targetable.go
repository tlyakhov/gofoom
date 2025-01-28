// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"bytes"
	"fmt"
	"text/template"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/ecs"

	"github.com/spf13/cast"
)

type PlayerTargetable struct {
	ecs.Attached `editable:"^"`
	Frob         core.Script `editable:"Frob"`
	Selected     core.Script `editable:"Selected"`
	UnSelected   core.Script `editable:"UnSelected"`
	Message      string      `editable:"Message"`

	MessageTemplate *template.Template
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

func (pt *PlayerTargetable) MultiAttachable() bool { return true }

func (pt *PlayerTargetable) Pos(e ecs.Entity) *concepts.Vector3 {
	if b := core.GetBody(pt.ECS, e); b != nil {
		top := &concepts.Vector3{}
		top[0] = b.Pos.Render[0]
		top[1] = b.Pos.Render[1]
		top[2] = b.Pos.Render[2] + b.Size.Render[1]*0.5
		return top
	} else if sector := core.GetSector(pt.ECS, e); sector != nil {
		return &sector.Center
	}
	return nil
}

func (pt *PlayerTargetable) String() string {
	return "PlayerTargetable"
}

func (pt *PlayerTargetable) OnAttach(db *ecs.ECS) {
	pt.Attached.OnAttach(db)
	pt.Frob.OnAttach(db)
	pt.Selected.OnAttach(db)
	pt.UnSelected.OnAttach(db)
}

type PlayerMessageParams struct {
	TargetableEntity ecs.Entity
	PlayerTargetable *PlayerTargetable
	Player           *Player
	InventoryCarrier *InventoryCarrier
}

func (pt PlayerTargetable) ApplyMessage(params *PlayerMessageParams) string {
	if pt.MessageTemplate == nil {
		return pt.Message
	}

	var buf bytes.Buffer
	err := pt.MessageTemplate.Execute(&buf, params)
	if err != nil {
		return fmt.Sprintf("Error in message template %v: %v", pt.Message, err)
	}
	return buf.String()
}

func (pt *PlayerTargetable) Construct(data map[string]any) {
	pt.Attached.Construct(data)

	if data == nil {
		pt.Frob.Construct(nil)
		pt.Selected.Construct(nil)
		pt.UnSelected.Construct(nil)
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
		pt.Message = cast.ToString(v)
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
