// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package core

import (
	"tlyakhov/gofoom/ecs"

	"github.com/spf13/cast"
)

type Scripted struct {
	ecs.AttachedWithIndirects `editable:"^"`

	OnFrame Script   `editable:"OnFrame"`
	Args    []string `editable:"Arguments"`
	Timer   float64  `editable:"ETA (timer)"` // ms

	TimerStart int64 // ns
}

func (s *Scripted) Shareable() bool { return true }

func (s *Scripted) Construct(data map[string]any) {
	s.AttachedWithIndirects.Construct(data)
	s.Args = nil
	s.Timer = 0
	s.TimerStart = ecs.Simulation.SimTimestamp

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

	if v, ok := data["Timer"]; ok {
		s.Timer = cast.ToFloat64(v)
	}
}

func (s *Scripted) Serialize() map[string]any {
	result := s.AttachedWithIndirects.Serialize()

	if !s.OnFrame.IsEmpty() {
		result["OnFrame"] = s.OnFrame.Serialize()
	}

	if len(s.Args) > 0 {
		result["Args"] = s.Args
	}

	if s.Timer > 0 {
		result["Timer"] = s.Timer
	}

	return result
}
