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

func (pt *PlayerTargetable) MultiAttachable() bool { return true }

func (pt *PlayerTargetable) Pos(e ecs.Entity) *concepts.Vector3 {
	if b := core.GetBody(e); b != nil {
		top := &concepts.Vector3{}
		top[0] = b.Pos.Render[0]
		top[1] = b.Pos.Render[1]
		top[2] = b.Pos.Render[2] + b.Size.Render[1]*0.5
		return top
	} else if sector := core.GetSector(e); sector != nil {
		return &sector.Center
	}
	return nil
}

func (pt *PlayerTargetable) String() string {
	if pt.MessageTemplate != nil {
		var buf bytes.Buffer
		err := pt.MessageTemplate.Execute(&buf, nil)
		if err != nil {
			return fmt.Sprintf("Error in message template %v: %v", pt.Message, err)
		}
		return buf.String()
	} else {
		return "PlayerTargetable"
	}

}

func (pt *PlayerTargetable) OnAttach() {
	pt.Attached.OnAttach()
	pt.Frob.OnAttach()
	pt.Selected.OnAttach()
	pt.UnSelected.OnAttach()
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
