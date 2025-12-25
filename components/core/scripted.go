// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package core

import (
	"tlyakhov/gofoom/ecs"

	"github.com/spf13/cast"
)

type Scripted struct {
	ecs.Attached      `editable:"^"`
	ecs.ApplyIndirect `editable:"^"`

	OnFrame Script   `editable:"OnFrame"`
	Args    []string `editable:"Arguments"`
}

func (s *Scripted) MultiAttachable() bool { return true }

func (s *Scripted) Construct(data map[string]any) {
	s.Attached.Construct(data)
	s.ApplyIndirect.Construct(data)

	if data == nil {
		s.OnFrame.Construct(nil)
		return
	}

	if v, ok := data["OnFrame"]; ok {
		s.OnFrame.Construct(v.(map[string]any))
	} else {
		s.OnFrame.Construct(nil)
	}

	if v, ok := data["Args"]; ok {
		s.Args = cast.ToStringSlice(v)
	}
}

func (s *Scripted) Serialize() map[string]any {
	result := s.Attached.Serialize()
	s.ApplyIndirect.Serialize(result)

	if !s.OnFrame.IsEmpty() {
		result["OnFrame"] = s.OnFrame.Serialize()
	}

	if len(s.Args) > 0 {
		result["Args"] = s.Args
	}

	return result
}
