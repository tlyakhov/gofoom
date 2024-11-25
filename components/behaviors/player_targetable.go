// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"bytes"
	"fmt"
	"text/template"
	"tlyakhov/gofoom/components/core"
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

func (pt *PlayerTargetable) String() string {
	return "PlayerTargetable"
}

func (pt *PlayerTargetable) AttachECS(db *ecs.ECS) {
	pt.Attached.AttachECS(db)
	pt.Frob.AttachECS(db)
	pt.Selected.AttachECS(db)
	pt.UnSelected.AttachECS(db)
}

func (pt PlayerTargetable) ApplyMessage(e ecs.Entity) string {
	if pt.MessageTemplate == nil {
		return pt.Message
	}

	var buf bytes.Buffer
	err := pt.MessageTemplate.Execute(&buf, e)
	if err != nil {
		return fmt.Sprintf("Error in message template %v: %v", pt.Message, err)
	}
	return buf.String()
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
