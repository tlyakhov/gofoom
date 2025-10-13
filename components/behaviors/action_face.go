// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"fmt"

	"github.com/spf13/cast"
)

// - Animated entities and NPCs can follow this
type ActionFace struct {
	ActionTimed `editable:"^"`

	Angle float64 `editable:"Angle"`
}

func (face *ActionFace) String() string {
	return fmt.Sprintf("Face: %.2f", face.Angle)
}

func (face *ActionFace) Construct(data map[string]any) {
	face.ActionTimed.Construct(data)
	face.Angle = 0

	if data == nil {
		return
	}

	if v, ok := data["Angle"]; ok {
		face.Angle = cast.ToFloat64(v)
	}
}

func (face *ActionFace) Serialize() map[string]any {
	result := face.ActionTimed.Serialize()
	result["Angle"] = face.Angle

	return result
}
